// cmd/websocket/main.go
package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/redis"
	"github.com/purushothdl/gochat-backend/internal/websocket"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Load application configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize structured JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create the Redis Presence Manager.
	presenceManager, err := redis.NewPresenceManager(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to create presence manager: %v", err)
	}

	// Create and start WebSocket hub
	hub := websocket.NewHub(logger, presenceManager)
	go hub.Run()

	// Register WebSocket endpoint handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, cfg, logger, w, r)
	})

	wsPort := os.Getenv("WEBSOCKET_PORT")
	if wsPort == "" {
		wsPort = "8081"
	}

	// Start WebSocket server
	logger.Info("websocket server starting", "port", wsPort)
	err = http.ListenAndServe(":"+wsPort, nil)
	if err != nil {
		log.Fatalf("Failed to start websocket server: %v", err)
	}
}