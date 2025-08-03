package room

import "context"

// UserProvider defines the contract for user-related checks needed by the room service.
type UserProvider interface {
	ExistsByID(ctx context.Context, id string) (bool, error)
}