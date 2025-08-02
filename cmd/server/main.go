package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/container"
	"github.com/purushothdl/gochat-backend/internal/database"
	httpTransport "github.com/purushothdl/gochat-backend/internal/transport/http"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("application startup error", "error", err)
		os.Exit(1)
	}
}

// run orchestrates the application startup, dependency injection, and server start.
func run(logger *slog.Logger) error {
	if err := godotenv.Load(); err != nil {
		logger.Warn(".env file not found, using environment variables")
	}

	// Load core application configuration.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Establish database connection and run migrations.
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Build the dependency injection container.
	c := container.New(cfg, db, logger)
	if err := c.Build(); err != nil {
		return fmt.Errorf("failed to build container: %w", err)
	}

	// Set up the HTTP router with all routes and middleware.
	router := httpTransport.NewRouter(c.AuthHandler, c.UserHandler, c.AuthMiddleware)
	handler := router.SetupRoutes(cfg, logger)

	// Create and run the server, which handles its own lifecycle.
	srv := httpTransport.NewServer(cfg, logger, handler)
	return srv.Run()
}