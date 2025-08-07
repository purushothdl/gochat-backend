// internal/websocket/protocol.go
package websocket

// Event is the structure for all messages sent over the WebSocket.
type Event struct {
	Type    string      `json:"type"`
	Payload any         `json:"payload"`
}