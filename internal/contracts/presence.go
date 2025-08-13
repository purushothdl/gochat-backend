package contracts

import "context"

// PresenceManager defines the contract for tracking user presence in rooms.
type PresenceManager interface {
	AddToRoom(ctx context.Context, roomID string, userID string) error
	RemoveFromRoom(ctx context.Context, roomID string, userID string) error
	GetOnlineUserIDs(ctx context.Context, roomID string) ([]string, error)
}