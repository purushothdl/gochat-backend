package room

import (
	"time"

	"github.com/google/uuid"
)

type RoomType string

const (
	DirectRoom  RoomType = "DIRECT"
	PrivateRoom RoomType = "PRIVATE"
	PublicRoom  RoomType = "PUBLIC"
)

type MemberRole string

const (
	AdminRole  MemberRole = "ADMIN"
	RegularRole MemberRole = "MEMBER"
)

// Room represents the core room entity.
type Room struct {
	ID              string
	Name            string
	Type            RoomType
	IsBroadcastOnly bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// RoomMembership links a user to a room with a specific role.
type RoomMembership struct {
	RoomID    string
	UserID    string
	Role      MemberRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPrivateRoom creates a new Room entity for a private chat.
func NewPrivateRoom(name string) *Room {
	return &Room{
		ID:   uuid.NewString(),
		Name: name,
		Type: PrivateRoom,
	}
}