package message

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	TypeText   MessageType = "TEXT"
	TypeSystem MessageType = "SYSTEM"
)

type Message struct {
	ID        string
	RoomID    string
	UserID    *string // Pointer to allow for NULL user (system messages)
	Content   string
	Type      MessageType
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewTextMessage creates a standard user-sent message entity.
func NewTextMessage(roomID, userID, content string) *Message {
	return &Message{
		ID:      uuid.NewString(),
		RoomID:  roomID,
		UserID:  &userID,
		Content: content,
		Type:    TypeText,
	}
}