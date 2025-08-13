// internal/websocket/hub.go
package websocket

import (
	"context"
	"log/slog"

	"github.com/purushothdl/gochat-backend/internal/contracts"
)

// SubscriptionRequest pairs a client with the rooms they want to join.
type SubscriptionRequest struct {
	Client  *Client
	RoomIDs []string
}

// Hub maintains the set of active clients and their room subscriptions.
type Hub struct {
	rooms      map[string]map[*Client]bool
	clients    map[*Client]bool
	subscribe  chan *SubscriptionRequest
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client

	logger     *slog.Logger
	presence   contracts.PresenceManager
}

func NewHub(logger *slog.Logger, presence contracts.PresenceManager) *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		clients:    make(map[*Client]bool),
		subscribe:  make(chan *SubscriptionRequest),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
		presence:   presence,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.Info("new client registered", "user_id", client.userID)

		case client := <-h.unregister:
			h.cleanupClient(client)

		case req := <-h.subscribe:
			h.handleSubscription(req)
		}
	}
}

// handleSubscription processes a client's request to join rooms.
func (h *Hub) handleSubscription(req *SubscriptionRequest) {
	for _, roomID := range req.RoomIDs {
		// If the room doesn't exist in the hub yet, create it.
		if _, ok := h.rooms[roomID]; !ok {
			h.rooms[roomID] = make(map[*Client]bool)
		}

		// Add the client to the room.
		h.rooms[roomID][req.Client] = true

		// Add the room to the client's own subscription set for easy cleanup.
		req.Client.rooms[roomID] = true

		// Update presence in Redis.
		err := h.presence.AddToRoom(context.Background(), roomID, req.Client.userID)
		if err != nil {
			h.logger.Error("failed to add user to room presence", "error", err, "user_id", req.Client.userID, "room_id", roomID)
		}

		h.logger.Info("client subscribed to room", "user_id", req.Client.userID, "room_id", roomID)
	}
}

// cleanupClient removes a client and all of its subscriptions.
func (h *Hub) cleanupClient(client *Client) {
	if _, ok := h.clients[client]; !ok {
		return 
	}

	// Remove the client from every room it was subscribed to.
	for roomID := range client.rooms {
		// Update presence in Redis.
		err := h.presence.RemoveFromRoom(context.Background(), roomID, client.userID)
		if err != nil {
			h.logger.Error("failed to remove user from room presence", "error", err, "user_id", client.userID, "room_id", roomID)
		}

		if room, ok := h.rooms[roomID]; ok {
			delete(room, client)
			// If the room is now empty, delete it from the hub to save memory.
			if len(room) == 0 {
				delete(h.rooms, roomID)
			}
		}
	}

	// Remove the client from the main registry and close its send channel.
	delete(h.clients, client)
	close(client.send)

	h.logger.Info("client unregistered and cleaned up", "user_id", client.userID)
}
