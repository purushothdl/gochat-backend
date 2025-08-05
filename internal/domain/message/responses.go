package message

import (
	"time"

	"github.com/purushothdl/gochat-backend/internal/shared/types"
)

type MessageResponse struct {
	ID              string             `json:"id"`
	RoomID          string             `json:"room_id"`
	Content         string             `json:"content"`
	Type            MessageType        `json:"type"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	IsEdited        bool               `json:"is_edited"`
	Sender          *types.BasicUser   `json:"sender,omitempty"`
}

// PaginatedMessagesResponse is the structured response for message history.
type PaginatedMessagesResponse struct {
	Data       []*MessageResponse `json:"data"`
	NextCursor *time.Time         `json:"next_cursor,omitempty"`
	HasMore    bool               `json:"has_more"`
}

func (m *MessageWithSeenFlag) ToResponse() *MessageResponse {
	// For soft-deleted messages, mask the content.
	if m.DeletedAt != nil {
		return &MessageResponse{
			ID:        m.ID,
			RoomID:    m.RoomID,
			Content:   "This message was deleted",
			Type:      TypeSystem,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			IsEdited:  m.UpdatedAt.After(m.CreatedAt.Add(5 * time.Second)), // Add buffer
			Sender:    nil, 
		}
	}
	
	return &MessageResponse{
		ID:        m.ID,
		RoomID:    m.RoomID,
		Content:   m.Content,
		Type:      m.Type,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		IsEdited:  m.UpdatedAt.After(m.CreatedAt.Add(5 * time.Second)),
		Sender:    m.User,
	}
}