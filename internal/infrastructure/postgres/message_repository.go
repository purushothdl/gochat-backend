// internal/infrastructure/postgres/message_repository.go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/purushothdl/gochat-backend/internal/domain/message"
	"github.com/purushothdl/gochat-backend/internal/shared/types"
)

type MessageRepository struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{pool: pool}
}

// ============================================================================
// Message Operations
// ============================================================================

func (r *MessageRepository) CreateMessage(ctx context.Context, msg *message.Message) error {
	query := `
        INSERT INTO messages (id, room_id, user_id, content, type)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING created_at, updated_at
    `
	return r.pool.QueryRow(ctx, query, msg.ID, msg.RoomID, msg.UserID, msg.Content, msg.Type).Scan(
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
}

func (r *MessageRepository) GetMessageByID(ctx context.Context, messageID string) (*message.Message, error) {
	query := `SELECT id, room_id, user_id, content, type, created_at, updated_at, deleted_at FROM messages WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, messageID)
	msg, err := scanMessage(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, message.ErrMessageNotFound
		}
		return nil, err
	}
	return msg, nil
}

func (r *MessageRepository) ListMessagesByRoom(ctx context.Context, roomID, userID string, cursor message.PaginationCursor) ([]*message.MessageWithSeenFlag, error) {
	query := `
        SELECT
            m.id, m.room_id, m.user_id, m.content, m.type, m.created_at, m.updated_at, m.deleted_at,
            CASE WHEN mr.message_id IS NOT NULL THEN TRUE ELSE FALSE END as is_seen_by_user,
            u.id as sender_id, u.name as sender_name, u.image_url as sender_image_url
        FROM messages m
        LEFT JOIN users u ON m.user_id = u.id
        LEFT JOIN message_read_receipts mr ON m.id = mr.message_id AND mr.user_id = $1
        LEFT JOIN user_message_deletions umd ON m.id = umd.message_id AND umd.user_id = $1
        WHERE m.room_id = $2
          AND m.created_at < $3
          AND umd.message_id IS NULL -- Filter out messages deleted for the user
        ORDER BY m.created_at DESC
        LIMIT $4
    `
	rows, err := r.pool.Query(ctx, query, userID, roomID, cursor.Timestamp, cursor.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []*message.MessageWithSeenFlag
	for rows.Next() {
		var msg message.MessageWithSeenFlag
		var sender types.BasicUser
		var senderID, senderName, senderImageURL pgtype.Text 

		err := rows.Scan(
			&msg.ID, &msg.RoomID, &msg.UserID, &msg.Content, &msg.Type, &msg.CreatedAt, &msg.UpdatedAt, &msg.DeletedAt,
			&msg.IsSeenByUser,
			&senderID, &senderName, &senderImageURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message with seen flag: %w", err)
		}

		if senderID.Valid {
			sender.ID = senderID.String
			sender.Name = senderName.String
			sender.ImageURL = senderImageURL.String
			msg.User = &sender
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, messageID, content string) error {
	query := `UPDATE messages SET content = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, content, messageID)
	return err
}

// ============================================================================
// Deletion Operations
// ============================================================================

func (r *MessageRepository) SoftDeleteMessage(ctx context.Context, messageID string) error {
	query := `UPDATE messages SET content = '', deleted_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, messageID)
	return err
}

func (r *MessageRepository) DeleteMessageForUser(ctx context.Context, messageID, userID string) error {
	query := `INSERT INTO user_message_deletions (message_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, messageID, userID)
	return err
}

// ============================================================================
// Receipt Operations
// ============================================================================

func (r *MessageRepository) UpdateRoomReadMarker(ctx context.Context, roomID, userID string, timestamp time.Time) error {
	query := `
        UPDATE room_memberships SET last_read_timestamp = $3
        WHERE room_id = $1 AND user_id = $2 AND (last_read_timestamp IS NULL OR last_read_timestamp < $3)
    `
	_, err := r.pool.Exec(ctx, query, roomID, userID, timestamp)
	return err
}

func (r *MessageRepository) CreateBulkReadReceipts(ctx context.Context, roomID, userID string, messageIDs []string) error {
	// pgx's large batch insert is the most efficient way to handle this.
	rows := make([][]interface{}, len(messageIDs))
	for i, msgID := range messageIDs {
		rows[i] = []interface{}{msgID, userID, roomID}
	}

	_, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"message_read_receipts"},
		[]string{"message_id", "user_id", "room_id"},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (r *MessageRepository) GetMessageReceipts(ctx context.Context, messageID string) ([]*types.BasicUser, error) {
	// This query now correctly selects only the columns needed for a BasicUser.
	query := `
        SELECT u.id, u.name, u.image_url
        FROM users u
        JOIN message_read_receipts mrr ON u.id = mrr.user_id
        WHERE mrr.message_id = $1 AND u.deleted_at IS NULL
    `
	rows, err := r.pool.Query(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message receipts: %w", err)
	}
	defer rows.Close()

	var users []*types.BasicUser
	for rows.Next() {
		user, err := scanBasicUserFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan receipt user: %w", err)
}
		users = append(users, user)
	}
	return users, nil
}

// ============================================================================
// Private Helpers
// ============================================================================

func scanMessage(row pgx.Row) (*message.Message, error) {
	var m message.Message
	err := row.Scan(&m.ID, &m.RoomID, &m.UserID, &m.Content, &m.Type, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
	return &m, err
}