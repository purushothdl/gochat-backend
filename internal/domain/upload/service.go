package upload

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/contracts"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/imageproc"
)

type Service struct {
	storage   contracts.FileStorage
	queue     contracts.Queue
	processor *imageproc.Processor
	config    *config.Config
	logger    *slog.Logger
}

func NewService(
	storage   contracts.FileStorage,
	queue     contracts.Queue,
	processor *imageproc.Processor,
	config    *config.Config,
	logger    *slog.Logger,
) *Service {
	return &Service{
		storage:   storage,
		queue:     queue,
		processor: processor,
		config:    config,
		logger:    logger,
	}
}

func (s *Service) InitiateProfileImageUpload(ctx context.Context, userID string, file multipart.File, header *multipart.FileHeader) (*JobResponse, error) {
	if header.Size > s.config.Upload.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds %d bytes", s.config.Upload.MaxFileSize)
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	if err := s.processor.ValidateImage(fileData); err != nil {
		return nil, fmt.Errorf("invalid image format: %w", err)
	}

	jobID := uuid.NewString()
	stagingKey := fmt.Sprintf("%s/%s_%s", s.config.Upload.StagingPath, jobID, header.Filename)

	contentType := header.Header.Get("Content-Type")
	if err := s.storage.Upload(ctx, stagingKey, contentType, bytes.NewReader(fileData), false); err != nil {
		return nil, fmt.Errorf("failed to upload to staging: %w", err)
	}

	job := ProfileUploadJob{
		JobID:        jobID,
		UserID:       userID,
		StagingKey:   stagingKey,
		OriginalName: header.Filename,
	}

	if err := s.queue.Enqueue(ctx, s.config.Upload.QueueName, job); err != nil {
		s.logger.Error("failed to enqueue processing job", "error", err, "job_id", jobID)
		// Attempt to clean up the staged file if queuing fails.
		s.storage.Delete(context.Background(), stagingKey)
		return nil, fmt.Errorf("failed to schedule image for processing: %w", err)
	}

	return &JobResponse{JobID: jobID}, nil
}