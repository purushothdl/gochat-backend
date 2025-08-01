package main

import (
    "log"
    "net/http"

    "github.com/joho/godotenv"
    "github.com/purushothdl/gochat-backend/internal/config"
    "github.com/purushothdl/gochat-backend/internal/container"
    "github.com/purushothdl/gochat-backend/internal/database"
    httpTransport "github.com/purushothdl/gochat-backend/internal/transport/http"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: .env file not found: %v", err)
    }

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Connect to database
    db, err := database.Connect(&cfg.Database)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    // Run migrations
    migrationRunner := database.NewMigrationRunner(db)
    if err := migrationRunner.RunMigrations(); err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }
    log.Println("Migrations completed successfully")

    // Build container with dependencies
    c := container.New(cfg, db)
    if err := c.Build(); err != nil {
        log.Fatalf("Failed to build container: %v", err)
    }

    // Setup router
    router := httpTransport.NewRouter(c.AuthHandler, c.UserHandler, c.AuthMiddleware)
    routes := router.SetupRoutes()

    // Start server
    log.Printf("Server starting on port %s", cfg.Server.Port)
    log.Printf("Health check: http://localhost:%s/api/health", cfg.Server.Port)
    
    if err := http.ListenAndServe(":"+cfg.Server.Port, routes); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
