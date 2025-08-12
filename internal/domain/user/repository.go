package user

import (
	"context"

	"github.com/purushothdl/gochat-backend/internal/shared/types"
)

type Repository interface {
    // User domain methods (full entity)
    ExistsByID(ctx context.Context, id string) (bool, error)
    GetByID(ctx context.Context, id string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
    UpdateUserImageURL(ctx context.Context, userID string, imageURL string) error

    // User Block Methods
	BlockUser(ctx context.Context, blockerID, blockedID string) error
	UnblockUser(ctx context.Context, blockerID, blockedID string) error
	ListBlockedUsers(ctx context.Context, blockerID string) ([]*types.BasicUser, error)
	IsBlocked(ctx context.Context, userID1, userID2 string) (bool, error)
}
