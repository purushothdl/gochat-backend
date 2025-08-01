package auth

import "context"

type Repository interface {
    // Refresh token management
    CreateRefreshToken(ctx context.Context, token *RefreshToken) error
    GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
    UpdateRefreshTokenUsage(ctx context.Context, tokenHash string) error
    DeleteRefreshToken(ctx context.Context, tokenHash string) error
    DeleteUserRefreshTokens(ctx context.Context, userID string) error
    CleanupExpiredTokens(ctx context.Context) error
    
    // User token management
    GetUserRefreshTokens(ctx context.Context, userID string) ([]*RefreshToken, error)
    DeleteOldestUserToken(ctx context.Context, userID string) error
    CountUserTokens(ctx context.Context, userID string) (int, error)
}
