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
	logger.Info("Incoming WebSocket request headers", "headers", r.Header)

	// 1. Authenticate the user from the access_token cookie.
	cookie, err := r.Cookie("access_token")
	if err != nil {
		logger.Error("Failed to get access_token cookie", "error", err) 
		http.Error(w, "Unauthorized: No access token", http.StatusUnauthorized)
		return
	}
	accessToken := cookie.Value
	logger.Info("Access token cookie found", "token_present", len(accessToken) > 0) 

	claims, err := auth.ValidateAccessToken(accessToken, cfg.JWT.Secret)
	if err != nil {
		logger.Error("Failed to validate access token", "error", err) 
		http.Error(w, "Unauthorized: Invalid access token", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID
	logger.Info("User authenticated via WebSocket", "user_id", userID) 

	// 2. Upgrade the HTTP connection to a WebSocket connection.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("failed to upgrade connection", "error", err)
		return
	}

	// 3. Create and register the new client.
	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		logger: logger,
	}
	client.hub.register <- client

	// 4. Start the read and write pumps in separate goroutines.
	go client.writePump()
	go client.readPump()
}
