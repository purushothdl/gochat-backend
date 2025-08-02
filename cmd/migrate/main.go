package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/database"
)

func main() {
	// Define flags
	up := flag.Bool("up", false, "Apply all available migrations")
	down := flag.Bool("down", false, "Rollback the last migration")
	version := flag.Bool("version", false, "Print the current schema version")
	force := flag.Int("force", 0, "Force version N, but no migration run") 
	flag.Parse()

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

	// Create migration runner
	migrationRunner := database.NewMigrationRunner(db)

	switch {
	case *up:
		log.Println("Applying migrations...")
		if err := migrationRunner.RunMigrations(); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		log.Println("Migrations applied successfully!")
	case *down:
		log.Println("Rolling back last migration...")
		if err := migrationRunner.RunDownMigration(); err != nil { 
			log.Fatalf("Failed to roll back migration: %v", err)
		}
		log.Println("Last migration rolled back successfully!")
	case *version:
		log.Println("Checking database version...")
		if currentVersion, dirty, err := migrationRunner.GetVersion(); err != nil { 
			log.Fatalf("Failed to get database version: %v", err)
		} else {
			log.Printf("Current database version: %d (Dirty: %t)\n", currentVersion, dirty)
		}
	case *force != 0: 
		log.Printf("Forcing database version to %d...\n", *force)
		if err := migrationRunner.ForceVersion(*force); err != nil { 
			log.Fatalf("Failed to force version: %v", err)
		}
		log.Printf("Database version forced to %d.\n", *force)
	default:
		fmt.Println("Please provide a migration command: -up, -down, -version, or -force N")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
