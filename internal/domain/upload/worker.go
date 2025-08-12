package upload

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/contracts"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/imageproc"
)

type Worker struct {
	queue        contracts.Queue
	storage      contracts.FileStorage
	userUpdater  UserProfileUpdater
	processor    *imageproc.Processor
	config       *config.Config
	logger       *slog.Logger
}

func NewWorker(
	queue        contracts.Queue,
	storage      contracts.FileStorage,
	userUpdater  UserProfileUpdater,
	processor    *imageproc.Processor,
	config       *config.Config,
	logger       *slog.Logger,
) *Worker {
	return &Worker{
		queue:        queue,
		storage:      storage,
		userUpdater:  userUpdater,
		processor:    processor,
		config:       config,
		logger:       logger,
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
		// If context is cancelled, redis returns an error, so we just return.
		if ctx.Err() != nil {
			return
		}
		w.logger.Error("failed to dequeue job", "error", err)
		// Sleep briefly to prevent fast-spinning loop on persistent redis error.
		time.Sleep(5 * time.Second)
		return
	}

	logger := w.logger.With("job_id", job.JobID, "user_id", job.UserID)
	logger.Info("processing new upload job")

	err := w.processUpload(&job, logger)
	// Always clean up the staging file, regardless of success or failure.
	defer func() {
		if cleanupErr := w.storage.Delete(context.Background(), job.StagingKey); cleanupErr != nil {
			logger.Error("failed to cleanup staging file", "error", cleanupErr, "key", job.StagingKey)
		}
	}()

	if err != nil {
		logger.Error("failed to process upload", "error", err)
		// Optionally, you could update the user's table with a processing error status.
		return
	}

	logger.Info("successfully processed upload job")
}

func (w *Worker) processUpload(job *ProfileUploadJob, logger *slog.Logger) error {
	stagedData, err := w.storage.Download(context.Background(), job.StagingKey)
	if err != nil {
		return fmt.Errorf("failed to download from staging: %w", err)
	}

	processedImage, err := w.processor.ProcessProfileImage(stagedData)
	if err != nil {
		return fmt.Errorf("failed to process image: %w", err)
	}

	// The final key is deterministic, based on user ID.
	finalKey := fmt.Sprintf("%s/%s.webp", w.config.Upload.ProfileImagePath, job.UserID)

	if err := w.storage.Upload(context.Background(), finalKey, processedImage.ContentType, bytes.NewReader(processedImage.Data), true); err != nil {
		return fmt.Errorf("failed to upload final image: %w", err)
	}

	imageURL := w.storage.GetPublicURL(finalKey)
	if err := w.userUpdater.UpdateUserImageURL(context.Background(), job.UserID, imageURL); err != nil {
		// If this fails, we have an orphan S3 file. This is a critical error to log.
		return fmt.Errorf("failed to update user profile with image url: %w", err)
	}

	// TODO: Publish a 'user:profile:updated' event to Redis Pub/Sub for WebSocket notification.
	logger.Info("user profile image updated", "url", imageURL)
	return nil
}