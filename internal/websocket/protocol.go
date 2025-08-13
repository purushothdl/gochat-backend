package websocket

import "encoding/json"

type EventType string

const (
	EventSubscribe EventType = "SUBSCRIBE"
	EventUnsubscribe EventType = "UNSUBSCRIBE"

	// TODO: Add other events like USER_TYPING, MESSAGE_READ etc.
)

// Event is the generic structure for all messages sent over the WebSocket.
type Event struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SubscribePayload is the specific payload for a SUBSCRIBE event.
type SubscribePayload struct {
	RoomIDs []string `json:"room_ids"`
}

// UnsubscribePayload is the specific payload for an UNSUBSCRIBE event.
type UnsubscribePayload struct {
	RoomIDs []string `json:"room_ids"`
}