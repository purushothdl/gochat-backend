package websocket

import (
	"context"
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
	ctx    context.Context
	cancel context.CancelFunc
}

// readPump parses messages from the client and sends them to the hub.
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

func (c *Client) handleNewMessage(message []byte) {
	var event Event
	if err := json.Unmarshal(message, &event); err != nil {
		c.logger.Error("failed to unmarshal event", "error", err)
		return
	}

	if event.Type == EventSubscribe {
		var payload SubscribePayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			c.logger.Error("failed to unmarshal subscribe payload", "error", err)
			return
		}
		// Pass the new payload structure to the hub.
		c.hub.subscribe <- &SubscriptionRequest{
			Client:       c,
			ChannelNames: payload.Channels,
		}
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.ctx.Done():
			// The context was canceled, indicating the client is being cleaned up.
			return
		}
	}
}