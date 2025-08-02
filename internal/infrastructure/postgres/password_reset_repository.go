package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/purushothdl/gochat-backend/internal/domain/auth"
)

type PasswordResetRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{pool: pool}
}

func (r *PasswordResetRepository) Store(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	query := `INSERT INTO password_reset_tokens (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`

	_, err := r.pool.Exec(ctx, query, tokenHash, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store password reset token: %w", err)
	}
	return nil
}

func (r *PasswordResetRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*auth.PasswordResetToken, error) {
	query := `SELECT user_id, expires_at FROM password_reset_tokens WHERE token_hash = $1`

	var t auth.PasswordResetToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(&t.UserID, &t.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to find password reset token: %w", err)
	}

	return &t, nil
}

func (r *PasswordResetRepository) Delete(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM password_reset_tokens WHERE token_hash = $1`

	_, err := r.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete password reset token: %w", err)
	}
	return nil
}