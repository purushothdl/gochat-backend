package user

import (
	"context"
	"mime/multipart"

	"github.com/purushothdl/gochat-backend/internal/domain/upload"
)

// External services user domain needs
type NotificationService interface {
    SendWelcomeEmail(ctx context.Context, userID string) error
    SendEmailVerification(ctx context.Context, userID, email string) error
    SendProfileUpdateNotification(ctx context.Context, userID string) error
}

type ProfileImageUploader interface {
	InitiateProfileImageUpload(ctx context.Context, userID string, file multipart.File, header *multipart.FileHeader) (*upload.JobResponse, error)
}