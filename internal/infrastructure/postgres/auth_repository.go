package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/purushothdl/gochat-backend/internal/domain/auth"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token *auth.RefreshToken) error {
	query := `
        INSERT INTO refresh_tokens (id, user_id, token_hash, device_info, ip_address, user_agent, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING created_at, last_used_at
    `

	err := r.pool.QueryRow(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.DeviceInfo,
		token.IPAddress, token.UserAgent, token.ExpiresAt,
	).Scan(&token.CreatedAt, &token.LastUsedAt)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*auth.RefreshToken, error) {
	query := `
        SELECT id, user_id, token_hash, device_info, ip_address::text, user_agent,
               expires_at, created_at, last_used_at
        FROM refresh_tokens 
        WHERE token_hash = $1 AND expires_at > NOW()
    `

	var token auth.RefreshToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.DeviceInfo,
		&token.IPAddress, &token.UserAgent, &token.ExpiresAt,
		&token.CreatedAt, &token.LastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &token, nil
}

func (r *AuthRepository) GetRefreshTokenByID(ctx context.Context, tokenID string) (*auth.RefreshToken, error) {
	query := `
        SELECT id, user_id, token_hash, device_info, 
               ip_address::text, user_agent,  
               expires_at, created_at, last_used_at
        FROM refresh_tokens 
        WHERE id = $1 AND expires_at > NOW()
    `

	var token auth.RefreshToken
	err := r.pool.QueryRow(ctx, query, tokenID).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.DeviceInfo,
		&token.IPAddress, &token.UserAgent, &token.ExpiresAt,
		&token.CreatedAt, &token.LastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token by ID: %w", err)
	}

	return &token, nil
}

func (r *AuthRepository) UpdateRefreshTokenUsage(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET last_used_at = NOW() WHERE token_hash = $1`

	_, err := r.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to update refresh token usage: %w", err)
	}

	return nil
}

func (r *AuthRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`

	_, err := r.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func (r *AuthRepository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}

	return nil
}

func (r *AuthRepository) CleanupExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at <= NOW()`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetUserRefreshTokens(ctx context.Context, userID string) ([]*auth.RefreshToken, error) {
	query := `
        SELECT id, user_id, token_hash, device_info, 
               ip_address::text, user_agent,
               expires_at, created_at, last_used_at
        FROM refresh_tokens 
        WHERE user_id = $1 AND expires_at > NOW()
        ORDER BY created_at DESC
    `

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user refresh tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*auth.RefreshToken
	for rows.Next() {
		var token auth.RefreshToken
		err := rows.Scan(
			&token.ID, &token.UserID, &token.TokenHash, &token.DeviceInfo,
			&token.IPAddress, &token.UserAgent, &token.ExpiresAt,
			&token.CreatedAt, &token.LastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan refresh token: %w", err)
		}

		tokens = append(tokens, &token)
	}

	return tokens, nil
}

func (r *AuthRepository) DeleteOldestUserToken(ctx context.Context, userID string) error {
	query := `
        DELETE FROM refresh_tokens 
        WHERE id = (
            SELECT id FROM refresh_tokens 
            WHERE user_id = $1 
            ORDER BY created_at ASC 
            LIMIT 1
        )
    `

	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete oldest user token: %w", err)
	}

	return nil
}

func (r *AuthRepository) CountUserTokens(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1 AND expires_at > NOW()`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user tokens: %w", err)
	}

	return count, nil
}

func (r *AuthRepository) DeleteRefreshTokenByID(ctx context.Context, tokenID string) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token by ID: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetActiveRefreshTokenByUserIDAndDeviceInfo(ctx context.Context, userID, deviceInfo string) (*auth.RefreshToken, error) {
	query := `
        SELECT id, user_id, token_hash, device_info, 
               ip_address::text, user_agent,
               expires_at, created_at, last_used_at
        FROM refresh_tokens 
        WHERE user_id = $1 AND device_info = $2 AND expires_at > NOW()
        ORDER BY created_at DESC
        LIMIT 1
    `

	var token auth.RefreshToken
	err := r.pool.QueryRow(ctx, query, userID, deviceInfo).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.DeviceInfo,
		&token.IPAddress, &token.UserAgent, &token.ExpiresAt,
		&token.CreatedAt, &token.LastUsedAt,
	)

	if err != nil {
		// If no rows are found, it's not an error but a nil token
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active refresh token by user ID and device info: %w", err)
	}

	return &token, nil
}
