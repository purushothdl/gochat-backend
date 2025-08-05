// internal/domain/message/interfaces.go
package message

import (
	"context"
	"github.com/purushothdl/gochat-backend/internal/shared/types"
)

// RoomProvider defines the methods the message service needs about rooms.
type RoomProvider interface {
	GetRoomInfo(ctx context.Context, roomID string) (*types.RoomInfo, error)
	GetMembershipInfo(ctx context.Context, roomID, userID string) (*types.MembershipInfo, error)
}

// UserProvider defines the methods the message service needs about users.
type UserProvider interface {
	IsBlocked(ctx context.Context, userID1, userID2 string) (bool, error)
}