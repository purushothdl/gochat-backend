package auth

import (
    "context"
    "github.com/purushothdl/gochat-backend/internal/shared/types"
)

// Limited user repository interface - only what auth needs
type UserRepository interface {
    GetByEmailShared(ctx context.Context, email string) (*types.User, error)
	GetByIDShared(ctx context.Context, id string) (*types.User, error)
    Create(ctx context.Context, userData *types.CreateUserData) (*types.User, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    UpdateLastLogin(ctx context.Context, userID string) error
    GetPasswordHash(ctx context.Context, userID string) (string, error)
}
// External services auth domain might need
type NotificationService interface {
    SendWelcomeEmail(ctx context.Context, userID, email string) error
    SendPasswordResetEmail(ctx context.Context, userID, email string) error
}
