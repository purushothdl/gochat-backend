// internal/websocket/client.go
package websocket

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
	logger *slog.Logger
	rooms  map[string]bool
}

// readPump pumps messages from the websocket connection to the hub for processing.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Warn("unexpected close error", "error", err)
			}
			break
		}
		c.handleNewMessage(message)
	}
}

// handleNewMessage parses incoming messages and routes them to the hub.
func (c *Client) handleNewMessage(message []byte) {
	var event Event
	if err := json.Unmarshal(message, &event); err != nil {
		c.logger.Error("failed to unmarshal event", "error", err)
		return
	}

	switch event.Type {
	case EventSubscribe:
		var payload SubscribePayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			c.logger.Error("failed to unmarshal subscribe payload", "error", err)
			return
		}
		c.hub.subscribe <- &SubscriptionRequest{
			Client:  c,
			RoomIDs: payload.RoomIDs,
		}

	// TODO: Add cases for other events like "USER_TYPING"

	default:
		c.logger.Warn("unknown event type received", "type", event.Type)
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}