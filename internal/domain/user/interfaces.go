package user

import "context"

// External services user domain needs
type NotificationService interface {
    SendWelcomeEmail(ctx context.Context, userID string) error
    SendEmailVerification(ctx context.Context, userID, email string) error
    SendProfileUpdateNotification(ctx context.Context, userID string) error
}
