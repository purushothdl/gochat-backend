// cmd/websocket/main.go
package main

import (
	"fmt"
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
	pubsubProvider, err := redis.NewPubSubProvider(&cfg.Redis)
	if err != nil {
		logger.Error("failed to create pubsub provider", "error", err)
		os.Exit(1)
	}
	presenceManager, err := redis.NewPresenceManager(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to create presence manager: %v", err)
	}

	// Create and start WebSocket hub
	hub := websocket.NewHub(logger, pubsubProvider, presenceManager)
	go hub.Run()

	// The Handler now has fewer dependencies.
	wsHandler := websocket.NewHandler(hub, cfg, logger)

	http.HandleFunc("/ws", wsHandler.ServeWs)

	wsPort := os.Getenv("WEBSOCKET_PORT")
	if wsPort == "" {
		wsPort = "8081"
	}

	logger.Info(fmt.Sprintf("websocket server starting on port %s", wsPort))
	if err := http.ListenAndServe(":"+wsPort, nil); err != nil {
		logger.Error("failed to start websocket server", "error", err)
		os.Exit(1)
	}
}