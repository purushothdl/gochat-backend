package upload

import (
	"bytes"
	"context"
	"encoding/json" 
	"fmt"
	"log/slog"
	"time"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/contracts"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/imageproc"
	"github.com/purushothdl/gochat-backend/internal/websocket" 
)

type Worker struct {
	queue       contracts.Queue
	storage     contracts.FileStorage
	userUpdater UserProfileUpdater
	processor   *imageproc.Processor
	config      *config.Config
	logger      *slog.Logger
	pubsub      contracts.PubSub 
}

func NewWorker(
	queue       contracts.Queue,
	storage     contracts.FileStorage,
	userUpdater UserProfileUpdater,
	processor   *imageproc.Processor,
	config      *config.Config,
	logger      *slog.Logger,
	pubsub      contracts.PubSub,
) *Worker {
	return &Worker{
		queue:       queue,
		storage:     storage,
		userUpdater: userUpdater,
		processor:   processor,
		config:      config,
		logger:      logger,
		pubsub:      pubsub,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("starting profile upload worker...")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("profile upload worker shutting down")
			return
		default:
			w.processNextJob(ctx)
		}
	}
}

func (w *Worker) processNextJob(ctx context.Context) {
	var job ProfileUploadJob
	if err := w.queue.Dequeue(ctx, w.config.Upload.QueueName, &job); err != nil {
		if ctx.Err() != nil {
			return
		}
		w.logger.Error("failed to dequeue job", "error", err)
		time.Sleep(5 * time.Second)
		return
	}

	logger := w.logger.With("job_id", job.JobID, "user_id", job.UserID)
	logger.Info("processing new upload job")

	defer func() {
		if cleanupErr := w.storage.Delete(context.Background(), job.StagingKey); cleanupErr != nil {
			logger.Error("failed to cleanup staging file", "error", cleanupErr, "key", job.StagingKey)
		}
	}()
	
	if err := w.processUpload(ctx, &job, logger); err != nil {
		logger.Error("failed to process upload", "error", err)
		return
	}

	logger.Info("successfully processed upload job")
}


func (w *Worker) processUpload(ctx context.Context, job *ProfileUploadJob, logger *slog.Logger) error {
	stagedData, err := w.storage.Download(ctx, job.StagingKey)
	if err != nil {
		return fmt.Errorf("failed to download from staging: %w", err)
	}

	processedImage, err := w.processor.ProcessProfileImage(stagedData)
	if err != nil {
		return fmt.Errorf("failed to process image: %w", err)
	}

	finalKey := fmt.Sprintf("%s/%s.webp", w.config.Upload.ProfileImagePath, job.UserID)

	if err := w.storage.Upload(ctx, finalKey, processedImage.ContentType, bytes.NewReader(processedImage.Data), true); err != nil {
		return fmt.Errorf("failed to upload final image: %w", err)
	}

	imageURL := w.storage.GetPublicURL(finalKey)
	if err := w.userUpdater.UpdateUserImageURL(ctx, job.UserID, imageURL); err != nil {
		return fmt.Errorf("failed to update user record: %w", err)
	}

	// Publish the notification event to Redis.
	w.publishProfileUpdate(job.UserID, imageURL)

	logger.Info("user profile image updated", "url", imageURL)
	return nil
}

func (w *Worker) publishProfileUpdate(userID, newImageURL string) {
	payload := websocket.ProfileUpdatedPayload{
		NewImageURL: newImageURL,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		w.logger.Error("failed to marshal profile update payload", "error", err, "user_id", userID)
		return
	}

	event := websocket.Event{
		Type:    websocket.EventProfileUpdated,
		Payload: payloadBytes,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		w.logger.Error("failed to marshal profile update event", "error", err, "user_id", userID)
		return
	}

	channel := fmt.Sprintf("user:%s", userID)
	if err := w.pubsub.Publish(context.Background(), channel, string(eventBytes)); err != nil {
		w.logger.Error("failed to publish profile update event", "error", err, "user_id", userID, "channel", channel)
	} else {
		w.logger.Info("published profile update event", "user_id", userID, "channel", channel)
	}
}