package websocket

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/pkg/auth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Configure properly for production.
	},
}

// Handler holds dependencies for serving WebSocket connections.
type Handler struct {
	hub    *Hub
	config *config.Config
	logger *slog.Logger
}

func NewHandler(hub *Hub, cfg *config.Config, logger *slog.Logger) *Handler {
	return &Handler{
		hub:    hub,
		config: cfg,
		logger: logger,
	}
}

// ServeWs handles the entire lifecycle of a WebSocket connection.
func (h *Handler) ServeWs(w http.ResponseWriter, r *http.Request) {
	accessToken := h.getAccessToken(r)
	if accessToken == "" {
		http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateAccessToken(accessToken, h.config.JWT.Secret)
	if err != nil {
		h.logger.Error("websocket auth failed: invalid token", "error", err)
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade connection", "error", err, "user_id", claims.UserID)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	client := &Client{
		hub:    h.hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: claims.UserID,
		logger: h.logger.With("user_id", claims.UserID),
		rooms:  make(map[string]bool),
		ctx:    ctx,
		cancel: cancel,
	}

	// Register the client with the hub and start its pumps.
	h.hub.register <- client
	go client.writePump()
	go client.readPump()
}

func (h *Handler) getAccessToken(r *http.Request) string {
	if cookie, err := r.Cookie("access_token"); err == nil {
		return cookie.Value
	}
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}
	return ""
}