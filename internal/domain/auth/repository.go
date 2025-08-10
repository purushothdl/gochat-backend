package auth

import (
	"context"
	"time"
)

type Repository interface {
    // Refresh token management
    CreateRefreshToken(ctx context.Context, token *RefreshToken) error
    GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
    GetRefreshTokenByID(ctx context.Context, tokenID string) (*RefreshToken, error)
    UpdateRefreshTokenUsage(ctx context.Context, tokenHash string) error
    DeleteRefreshToken(ctx context.Context, tokenHash string) error
    DeleteRefreshTokenByID(ctx context.Context, tokenID string) error
    DeleteUserRefreshTokens(ctx context.Context, userID string) error
    CleanupExpiredTokens(ctx context.Context) error
    
    // User token management
    GetUserRefreshTokens(ctx context.Context, userID string) ([]*RefreshToken, error)
    DeleteOldestUserToken(ctx context.Context, userID string) error
    CountUserTokens(ctx context.Context, userID string) (int, error)
}

type PasswordResetRepository interface {
	Store(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	Delete(ctx context.Context, tokenHash string) error
}
