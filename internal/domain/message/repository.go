package message

import (
	"context"
	"time"

	"github.com/purushothdl/gochat-backend/internal/shared/types"
)

type Repository interface {
	CreateMessage(ctx context.Context, msg *Message) error
	GetMessageByID(ctx context.Context, messageID string) (*Message, error)
	ListMessagesByRoom(ctx context.Context, roomID, userID string, cursor PaginationCursor) ([]*MessageWithSeenFlag, error)
	UpdateMessage(ctx context.Context, messageID, content string) error

	SoftDeleteMessage(ctx context.Context, messageID string) error
	DeleteMessageForUser(ctx context.Context, messageID, userID string) error

	GetLatestTimestampForMessages(ctx context.Context, messageIDs []string) (*time.Time, error)
	UpdateRoomReadMarker(ctx context.Context, roomID, userID string, timestamp time.Time) error
	CreateBulkReadReceipts(ctx context.Context, roomID, userID string, messageIDs []string) error
	GetMessageReceipts(ctx context.Context, messageID string) ([]*types.ReceiptInfo, error) 
}