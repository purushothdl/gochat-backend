package websocket

import "encoding/json"

type EventType string

const (
	EventSubscribe EventType = "SUBSCRIBE"
	EventUnsubscribe EventType = "UNSUBSCRIBE"
	EventProfileUpdated EventType = "PROFILE_UPDATED"

	// TODO: Add other events like USER_TYPING, MESSAGE_READ etc.
)

// Event is the generic structure for all messages sent over the WebSocket.
type Event struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SubscribePayload is the specific payload for a SUBSCRIBE event.
type SubscribePayload struct {
	Channels []string `json:"channels"` 
}

// UnsubscribePayload is the specific payload for an UNSUBSCRIBE event.
type UnsubscribePayload struct {
	RoomIDs []string `json:"room_ids"`
}

// ProfileUpdatedPayload is the payload for the PROFILE_UPDATED event.
type ProfileUpdatedPayload struct {
	NewImageURL string `json:"new_image_url"`
}