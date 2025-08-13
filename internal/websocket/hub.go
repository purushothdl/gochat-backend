package websocket

import (
	"context"
	"log/slog"
	"strings"

	"github.com/purushothdl/gochat-backend/internal/contracts"
)

// SubscriptionRequest pairs a client with the channel names they want to join.
type SubscriptionRequest struct {
	Client       *Client
	ChannelNames []string
}

// Hub maintains the set of active clients and orchestrates subscriptions.
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	subscribe  chan *SubscriptionRequest
	logger     *slog.Logger
	pubsub     contracts.PubSub
	presence   contracts.PresenceManager
}

func NewHub(
	logger *slog.Logger,
	pubsub contracts.PubSub,
	presence contracts.PresenceManager,
) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		subscribe:  make(chan *SubscriptionRequest),
		logger:     logger,
		pubsub:     pubsub,
		presence:   presence,
	}
}

// Run starts the Hub's event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.Info("client registered locally", "user_id", client.userID)

		case client := <-h.unregister:
			h.cleanupClient(client)

		case req := <-h.subscribe:
			go h.startRedisSubscription(req)
		}
	}
}

// startRedisSubscription runs in a dedicated goroutine for each client's subscription set.
func (h *Hub) startRedisSubscription(req *SubscriptionRequest) {
	client := req.Client

	if len(req.ChannelNames) == 0 {
		return
	}

	for _, channelName := range req.ChannelNames {
		client.rooms[channelName] = true
		if id, ok := parseRoomID(channelName); ok {
			h.presence.AddToRoom(client.ctx, id, client.userID)
		}
	}

	messageChannel, err := h.pubsub.Subscribe(client.ctx, req.ChannelNames...)
	if err != nil {
		h.logger.Error("failed to subscribe to redis channels", "error", err, "user_id", client.userID)
		client.conn.Close()
		return
	}

	h.logger.Info("client subscribed to redis channels", "user_id", client.userID, "channels", req.ChannelNames)

	// This loop forwards messages from Redis to the client's send channel.
	for {
		select {
		case msg := <-messageChannel:
			if msg == nil {
				return
			}
			client.send <- []byte(msg.Payload)
		case <-client.ctx.Done():
			// The client's context was cancelled, so we exit this goroutine.
			return
		}
	}
}

// cleanupClient handles unregistering a client and cleaning their presence.
func (h *Hub) cleanupClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		client.cancel() 
		for channelName := range client.rooms {
			if id, ok := parseRoomID(channelName); ok {
				h.presence.RemoveFromRoom(context.Background(), id, client.userID)
			}
		}

		delete(h.clients, client)
		close(client.send)
		h.logger.Info("client unregistered and cleaned up", "user_id", client.userID)
	}
}

func parseRoomID(channelName string) (string, bool) {
	prefix := "room:"
	if strings.HasPrefix(channelName, prefix) {
		return strings.TrimPrefix(channelName, prefix), true
	}
	return "", false
}