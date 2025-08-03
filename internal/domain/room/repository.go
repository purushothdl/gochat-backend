package room

import "context"

// Repository defines the persistence interface for room and membership data.
type Repository interface {
	CreateRoom(ctx context.Context, room *Room) error
	FindRoomByID(ctx context.Context, roomID string) (*Room, error)
	ListPublicRooms(ctx context.Context) ([]*Room, error)
	UpdateRoom(ctx context.Context, room *Room) error

	CreateMembership(ctx context.Context, membership *RoomMembership) error
	FindMembership(ctx context.Context, roomID, userID string) (*RoomMembership, error)
	UpdateMembership(ctx context.Context, membership *RoomMembership) error
	DeleteMembership(ctx context.Context, roomID, userID string) error
	ListUserRooms(ctx context.Context, userID string) ([]*Room, error)
	ListMembers(ctx context.Context, roomID string) ([]*MemberDetail, error)
	CountAdmins(ctx context.Context, roomID string) (int, error) 
}