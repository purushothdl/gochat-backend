package message

import "time"

type CreateMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=2000"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=2000"`
}

type BulkSeenRequest struct {
	RoomID     string   `json:"room_id" validate:"required,uuid"`
	MessageIDs []string `json:"message_ids" validate:"required,min=1,dive,uuid"`
}

type ReadMarkerRequest struct {
	LastReadTimestamp time.Time `json:"last_read_timestamp" validate:"required"`
}