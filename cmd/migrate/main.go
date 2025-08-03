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
    create := flag.Bool("create", false, "Create a new migration file")
    list := flag.Bool("list", false, "List all available migrations")
    name := flag.String("name", "", "Migration name (for create action)")
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

    case *create:
        migrationName := *name
        if migrationName == "" {
            fmt.Print("Enter migration name: ")
            fmt.Scanln(&migrationName)
        }
        if err := migrationRunner.CreateMigration(migrationName); err != nil {
            log.Fatalf("Failed to create migration: %v", err)
        }

    case *list:
        if err := migrationRunner.ListMigrations(); err != nil {
            log.Fatalf("Failed to list migrations: %v", err)
        }

    default:
        fmt.Println("Please provide a migration command:")
        fmt.Println("  -up      Apply all available migrations")
        fmt.Println("  -down    Rollback the last migration")
        fmt.Println("  -version Print the current schema version")
        fmt.Println("  -force N Force version N")
        fmt.Println("  -create  Create a new migration file")
        fmt.Println("  -list    List all available migrations")
        fmt.Println("\nFor create command, optionally use -name flag:")
        fmt.Println("  ./migrate -create -name=create_users_table")
        os.Exit(1)
    }
}
