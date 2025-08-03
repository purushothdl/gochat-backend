// internal/infrastructure/postgres/room_repository.go
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/purushothdl/gochat-backend/internal/domain/room"
)

type RoomRepository struct {
	pool *pgxpool.Pool
}

func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

// CreateRoom inserts a new room record into the database and populates
// the created_at and updated_at fields of the Room object.
func (r *RoomRepository) CreateRoom(ctx context.Context, newRoom *room.Room) error {
	query := `
        INSERT INTO rooms (id, name, type, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        RETURNING created_at, updated_at
    `
	err := r.pool.QueryRow(ctx, query, newRoom.ID, newRoom.Name, newRoom.Type).Scan(
		&newRoom.CreatedAt,
		&newRoom.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}
	return nil
}

// UpdateRoom updates a room's mutable properties.
func (r *RoomRepository) UpdateRoom(ctx context.Context, rm *room.Room) error {
	query := `
        UPDATE rooms
        SET name = $2, is_broadcast_only = $3, updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	cmdTag, err := r.pool.Exec(ctx, query, rm.ID, rm.Name, rm.IsBroadcastOnly)
	if err != nil {
		return fmt.Errorf("failed to update room: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return room.ErrRoomNotFound
	}
	return nil
}

// CreateMembership inserts a new room membership record.
func (r *RoomRepository) CreateMembership(ctx context.Context, membership *room.RoomMembership) error {
	query := `
        INSERT INTO room_memberships (room_id, user_id, role, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
    `
	_, err := r.pool.Exec(ctx, query, membership.RoomID, membership.UserID, membership.Role)
	if err != nil {
		return fmt.Errorf("failed to create membership: %w", err)
	}
	return nil
}

// FindMembership retrieves a specific membership from the database.
func (r *RoomRepository) FindMembership(ctx context.Context, roomID, userID string) (*room.RoomMembership, error) {
	query := `SELECT room_id, user_id, role, created_at, updated_at FROM room_memberships WHERE room_id = $1 AND user_id = $2`
	var m room.RoomMembership
	err := r.pool.QueryRow(ctx, query, roomID, userID).Scan(
		&m.RoomID,
		&m.UserID,
		&m.Role,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, room.ErrNotMember
		}
		return nil, fmt.Errorf("failed to find membership: %w", err)
	}
	return &m, nil
}

// ListUserRooms retrieves all rooms a user is a member of.
func (r *RoomRepository) ListUserRooms(ctx context.Context, userID string) ([]*room.Room, error) {
	query := `
        SELECT r.id, r.name, r.type, r.created_at, r.updated_at
        FROM rooms r
        JOIN room_memberships rm ON r.id = rm.room_id
        WHERE rm.user_id = $1 AND r.deleted_at IS NULL
        ORDER BY r.updated_at DESC
    `
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*room.Room
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// FindRoomByID retrieves a single room by its ID.
func (r *RoomRepository) FindRoomByID(ctx context.Context, roomID string) (*room.Room, error) {
	query := `SELECT id, name, type, created_at, updated_at FROM rooms WHERE id = $1 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, query, roomID)
	foundRoom, err := scanRoom(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, room.ErrRoomNotFound
		}
		return nil, fmt.Errorf("failed to find room by id: %w", err)
	}
	return foundRoom, nil
}

// ListPublicRooms retrieves all rooms with type 'PUBLIC'.
func (r *RoomRepository) ListPublicRooms(ctx context.Context) ([]*room.Room, error) {
	query := `SELECT id, name, type, created_at, updated_at FROM rooms WHERE type = 'PUBLIC' AND deleted_at IS NULL ORDER BY updated_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list public rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*room.Room
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// ListMembers retrieves all members of a specific room, along with their public user details.
func (r *RoomRepository) ListMembers(ctx context.Context, roomID string) ([]*room.MemberDetail, error) {
	query := `
        SELECT rm.room_id, rm.user_id, rm.role, u.name, u.image_url
        FROM room_memberships rm
        JOIN users u ON rm.user_id = u.id
        WHERE rm.room_id = $1 AND u.deleted_at IS NULL
    `
	rows, err := r.pool.Query(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to list room members: %w", err)
	}
	defer rows.Close()

	var members []*room.MemberDetail
	for rows.Next() {
		var m room.MemberDetail
		var imageURL *string 

		if err := rows.Scan(&m.RoomID, &m.UserID, &m.Role, &m.Name, &imageURL); err != nil {
			return nil, fmt.Errorf("failed to scan member detail: %w", err)
		}

		if imageURL != nil {
			m.ImageURL = *imageURL
		}
		members = append(members, &m)
	}
	return members, nil
}

// UpdateMembership updates the role of a member in a room.
func (r *RoomRepository) UpdateMembership(ctx context.Context, membership *room.RoomMembership) error {
	query := `
        UPDATE room_memberships
        SET role = $3, updated_at = NOW()
        WHERE room_id = $1 AND user_id = $2
    `
	_, err := r.pool.Exec(ctx, query, membership.RoomID, membership.UserID, membership.Role)
	if err != nil {
		return fmt.Errorf("failed to update membership: %w", err)
	}
	return nil
}

// DeleteMembership removes a user from a room.
func (r *RoomRepository) DeleteMembership(ctx context.Context, roomID, userID string) error {
	query := `DELETE FROM room_memberships WHERE room_id = $1 AND user_id = $2`
	cmdTag, err := r.pool.Exec(ctx, query, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return room.ErrNotMember // The user wasn't in the room to begin with.
	}
	return nil
}

// CountAdmins counts the number of administrators in a given room.
func (r *RoomRepository) CountAdmins(ctx context.Context, roomID string) (int, error) {
	query := `SELECT COUNT(*) FROM room_memberships WHERE room_id = $1 AND role = 'ADMIN'`
	var count int
	err := r.pool.QueryRow(ctx, query, roomID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count admins: %w", err)
	}
	return count, nil
}

// scanRoom is a helper to scan a room record from a pgx.Row scanner.
func scanRoom(row pgx.Row) (*room.Room, error) {
	var r room.Room
	err := row.Scan(
		&r.ID,
		&r.Name,
		&r.Type,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan room: %w", err)
	}
	return &r, nil
}
