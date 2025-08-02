package auth

import (
	"context"

	"github.com/purushothdl/gochat-backend/internal/shared/types"
)

// Limited user repository interface - only what auth needs
type UserRepository interface {
	// User retrieval
	GetByEmailShared(ctx context.Context, email string) (*types.User, error)
	GetByIDShared(ctx context.Context, id string) (*types.User, error)
	
	// User creation
	Create(ctx context.Context, userData *types.CreateUserData) (*types.User, error)
	
	// User verification
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	
	// Authentication related
	GetPasswordHash(ctx context.Context, userID string) (string, error)
	UpdatePassword(ctx context.Context, userID string, newPasswordHash string) error
	
	// User activity tracking
	UpdateLastLogin(ctx context.Context, userID string) error
}

// External services auth domain might need
type NotificationService interface {
    SendWelcomeEmail(ctx context.Context, userID, email string) error
    SendPasswordResetEmail(ctx context.Context, userID, email string) error
}

type EmailService interface {
	SendPasswordResetEmail(ctx context.Context, recipientEmail, recipientName, resetLink string) error
}