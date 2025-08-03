// infrastructure/postgres/user_repository.go
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/purushothdl/gochat-backend/internal/domain/user"
	"github.com/purushothdl/gochat-backend/internal/shared/types"
)

type UserRepository struct {
    pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{pool: pool}
}

// ============================================================================
// USER DOMAIN METHODS (Returns user.User entities)
// ============================================================================

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
    query := `
        SELECT id, email, name, image_url, password_hash, settings, 
               created_at, updated_at, is_verified, last_login
        FROM users 
        WHERE id = $1 AND deleted_at IS NULL
    `
    
    return r.scanDomainUser(ctx, query, id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
    query := `
        SELECT id, email, name, image_url, password_hash, settings, 
               created_at, updated_at, is_verified, last_login
        FROM users 
        WHERE email = $1 AND deleted_at IS NULL
    `
    
    return r.scanDomainUser(ctx, query, email)
}


func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	settingsJSON, err := json.Marshal(u.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal user settings: %w", err)
	}

	query := `
		UPDATE users SET
			name = $2,
			image_url = $3,
			password_hash = $4,
			settings = $5
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err = r.pool.Exec(ctx, query, u.ID, u.Name, u.ImageURL, u.PasswordHash, settingsJSON)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
    query := `UPDATE users SET deleted_at = NOW() WHERE id = $1`
    _, err := r.pool.Exec(ctx, query, id)
    return err
}

// ============================================================================
// OTHER DOMAIN METHODS (Returns types.User - shared types)
// ============================================================================

func (r *UserRepository) GetByEmailShared(ctx context.Context, email string) (*types.User, error) {
    query := `
        SELECT id, email, name, image_url, created_at, updated_at, is_verified, last_login
        FROM users 
        WHERE email = $1 AND deleted_at IS NULL
    `
    
    return r.scanSharedUser(ctx, query, email)
}

func (r *UserRepository) GetByIDShared(ctx context.Context, id string) (*types.User, error) {
    query := `
        SELECT id, email, name, image_url, created_at, updated_at, is_verified, last_login
        FROM users 
        WHERE id = $1 AND deleted_at IS NULL
    `
    
    return r.scanSharedUser(ctx, query, id)
}

func (r *UserRepository) Create(ctx context.Context, userData *types.CreateUserData) (*types.User, error) {
    userID := uuid.New().String()

    defaultSettings := user.NewDefaultUserSettings()
    settingsJSON, err := json.Marshal(defaultSettings)
    if err != nil {
        // This is an internal error, should not happen with a valid struct
        return nil, fmt.Errorf("failed to marshal default settings: %w", err)
    }

    query := `
        INSERT INTO users (id, email, name, password_hash, is_verified, settings)
        VALUES ($1, $2, $3, $4, false, $5) -- Use a parameter for settings
        RETURNING id, email, name, image_url, created_at, updated_at, is_verified
    `
    
    var u types.User
    var imageURL sql.NullString  
    
    err = r.pool.QueryRow(ctx, query, userID, userData.Email, userData.Name, userData.Password, settingsJSON).Scan(
        &u.ID, &u.Email, &u.Name, &imageURL, &u.CreatedAt, &u.UpdatedAt, &u.IsVerified,
    )
    
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    // Handle NULL image_url
    if imageURL.Valid {
        u.ImageURL = imageURL.String
    } else {
        u.ImageURL = ""
    }
    
    return &u, nil
}

func (r *UserRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`
    
    var exists bool
    err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
    return exists, err
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
    
    var exists bool
    err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
    return exists, err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
    query := `UPDATE users SET last_login = NOW() WHERE id = $1`
    _, err := r.pool.Exec(ctx, query, userID)
    return err
}

func (r *UserRepository) GetPasswordHash(ctx context.Context, userID string) (string, error) {
    query := `SELECT password_hash FROM users WHERE id = $1 AND deleted_at IS NULL`
    
    var hash string
    err := r.pool.QueryRow(ctx, query, userID).Scan(&hash)
    return hash, err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID string, newPasswordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.pool.Exec(ctx, query, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to execute password update: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found for password update")
	}

	return nil
}

// BlockUser creates a new record in the user_blocks table.
func (r *UserRepository) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	query := `INSERT INTO user_blocks (blocker_id, blocked_id) VALUES ($1, $2)`
	_, err := r.pool.Exec(ctx, query, blockerID, blockedID)
	if err != nil {
		// pgerr.Code "23505" is the unique_violation for the primary key.
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return user.ErrAlreadyBlocked
		}
		return fmt.Errorf("failed to block user: %w", err)
	}
	return nil
}

// UnblockUser removes a record from the user_blocks table.
func (r *UserRepository) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	query := `DELETE FROM user_blocks WHERE blocker_id = $1 AND blocked_id = $2`
	cmdTag, err := r.pool.Exec(ctx, query, blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return user.ErrNotBlocked
	}
	return nil
}

// ListBlockedUsers retrieves all users that a specific user has blocked.
func (r *UserRepository) ListBlockedUsers(ctx context.Context, blockerID string) ([]*types.BasicUser, error) {
	query := `
        SELECT u.id, u.name, u.image_url
        FROM users u
        JOIN user_blocks ub ON u.id = ub.blocked_id
        WHERE ub.blocker_id = $1 AND u.deleted_at IS NULL
    `
	rows, err := r.pool.Query(ctx, query, blockerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list blocked users: %w", err)
	}
	defer rows.Close()

	var users []*types.BasicUser
	for rows.Next() {
		// Use the new, dedicated helper for BasicUser.
		user, err := scanBasicUserFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blocked user row: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// IsBlocked checks if a communication block exists between two users, in either direction.
func (r *UserRepository) IsBlocked(ctx context.Context, userID1, userID2 string) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 FROM user_blocks 
            WHERE (blocker_id = $1 AND blocked_id = $2) 
               OR (blocker_id = $2 AND blocked_id = $1)
        )
    `
	var isBlocked bool
	err := r.pool.QueryRow(ctx, query, userID1, userID2).Scan(&isBlocked)
	if err != nil {
		return false, fmt.Errorf("failed to check if blocked: %w", err)
	}
	return isBlocked, nil
}

// ============================================================================
// PRIVATE HELPER METHODS
// ============================================================================

// This allows us to use one scanning helper for both single and multi-row queries.
type rowScanner interface {
	Scan(dest ...any) error
}

func (r *UserRepository) scanDomainUser(ctx context.Context, query string, args ...any) (*user.User, error) {
    var u user.User
    var settingsJSON []byte
    var imageURL sql.NullString  
    var lastLogin sql.NullTime
    
    err := r.pool.QueryRow(ctx, query, args...).Scan(
        &u.ID, &u.Email, &u.Name, &imageURL, &u.PasswordHash,  
        &settingsJSON, &u.CreatedAt, &u.UpdatedAt, &u.IsVerified, &lastLogin,
    )
    
    if err != nil {
        return nil, err
    }
    
    // Handle NULL image_url
    if imageURL.Valid {
        u.ImageURL = imageURL.String
    } else {
        u.ImageURL = ""
    }
    
    // Parse settings JSON
    if len(settingsJSON) > 0 {
        json.Unmarshal(settingsJSON, &u.Settings)
    } else {
        u.Settings = user.UserSettings{
            Theme:                "light",
            NotificationsEnabled: true,
            Language:            "en",
        }
    }
    
    if lastLogin.Valid {
        u.LastLogin = &lastLogin.Time
    }
    
    return &u, nil
}

func (r *UserRepository) scanSharedUser(ctx context.Context, query string, args ...any) (*types.User, error) {
    var u types.User
    var imageURL sql.NullString  
    var lastLogin sql.NullTime

    err := r.pool.QueryRow(ctx, query, args...).Scan(
        &u.ID, &u.Email, &u.Name, &imageURL,  
        &u.CreatedAt, &u.UpdatedAt, &u.IsVerified, &lastLogin,
    )

    if err != nil {
        return nil, err
    }

    // Handle NULL image_url
    if imageURL.Valid {
        u.ImageURL = imageURL.String
    } else {
        u.ImageURL = "" 
    }

    // Handle NULL last_login
    if lastLogin.Valid {
        u.LastLogin = &lastLogin.Time
    }

    return &u, nil
}

// scanBasicUserFromRow scans the essential public fields of a user into a types.BasicUser.
func scanBasicUserFromRow(row rowScanner) (*types.BasicUser, error) {
	var u types.BasicUser
	var imageURL sql.NullString

	err := row.Scan(&u.ID, &u.Name, &imageURL)
	if err != nil {
		return nil, err
	}

	if imageURL.Valid {
		u.ImageURL = imageURL.String
	}

	return &u, nil
}