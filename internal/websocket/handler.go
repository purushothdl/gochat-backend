// internal/websocket/handler.go
package websocket

import (
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
		// This should be configured properly for production.
		return true
	},
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, cfg *config.Config, logger *slog.Logger, w http.ResponseWriter, r *http.Request) {
	var accessToken string

	// 1. Try to get the token from the cookie first (for browsers).
	cookie, err := r.Cookie("access_token")
	if err == nil {
		accessToken = cookie.Value
	}

	// 2. If no cookie, fall back to checking the query parameter (for non-browser clients).
	if accessToken == "" {
		accessToken = r.URL.Query().Get("token")
	}

	// 3. If still no token, fail the connection.
	if accessToken == "" {
		logger.Error("websocket auth failed: no token provided in cookie or query param")
		http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
		return
	}

	// 4. Validate the token.
	claims, err := auth.ValidateAccessToken(accessToken, cfg.JWT.Secret)
	if err != nil {
		logger.Error("websocket auth failed: invalid token", "error", err)
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// 5. Upgrade the HTTP connection to a WebSocket connection.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("failed to upgrade connection", "error", err)
		return
	}

	// 6. Create and register the new client.
	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		logger: logger.With("user_id", userID),
		rooms:  make(map[string]bool),
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}