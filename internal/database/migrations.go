package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/postgres/migrations"
)

type MigrationRunner struct {
    pool         *pgxpool.Pool
    migrationDir string 
}

func NewMigrationRunner(pool *pgxpool.Pool) *MigrationRunner {
    return &MigrationRunner{
        pool:         pool,
        migrationDir: "internal/infrastructure/postgres/migrations", 
    }
}

// SetMigrationDir allows customizing the migration directory
func (m *MigrationRunner) SetMigrationDir(dir string) {
    m.migrationDir = dir
}

// CreateMigration creates new up and down migration files
func (m *MigrationRunner) CreateMigration(name string) error {
    if strings.TrimSpace(name) == "" {
        return fmt.Errorf("migration name cannot be empty")
    }

    // Sanitize the name (remove spaces, special chars)
    name = strings.ReplaceAll(strings.ToLower(name), " ", "_")
    name = strings.ReplaceAll(name, "-", "_")

    // Create directory if it doesn't exist
    if err := os.MkdirAll(m.migrationDir, 0755); err != nil {
        return fmt.Errorf("failed to create migration directory: %w", err)
    }

    timestamp := time.Now().Format("20060102150405")
    upFile := filepath.Join(m.migrationDir, fmt.Sprintf("%s_%s.up.sql", timestamp, name))
    downFile := filepath.Join(m.migrationDir, fmt.Sprintf("%s_%s.down.sql", timestamp, name))

    upContent := fmt.Sprintf(`-- Migration: %s
-- Created at: %s
-- Description: Add your migration description here

-- Add your UP migration SQL here
-- Example:
-- CREATE TABLE users (
--     id SERIAL PRIMARY KEY,
--     username VARCHAR(255) NOT NULL UNIQUE,
--     email VARCHAR(255) NOT NULL UNIQUE,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

`, name, time.Now().Format(time.RFC3339))

    downContent := fmt.Sprintf(`-- Rollback migration: %s
-- Created at: %s
-- Description: Rollback changes from %s migration

-- Add your DOWN migration SQL here
-- Example:
-- DROP TABLE IF EXISTS users;

`, name, time.Now().Format(time.RFC3339), name)

    // Write the files
    if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
        return fmt.Errorf("failed to create up migration file: %w", err)
    }

    if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
        return fmt.Errorf("failed to create down migration file: %w", err)
    }

    fmt.Printf("Created migration files:\n")
    fmt.Printf("   %s\n", upFile)
    fmt.Printf("   %s\n", downFile)
    fmt.Printf("\nNext steps:\n")
    fmt.Printf("   1. Edit the .up.sql file with your migration SQL\n")
    fmt.Printf("   2. Edit the .down.sql file with rollback SQL\n")
    fmt.Printf("   3. Run 'make migrate-up' to apply the migration\n")

    return nil
}

// ListMigrations shows all available migration files
func (m *MigrationRunner) ListMigrations() error {
    files, err := os.ReadDir(m.migrationDir)
    if err != nil {
        return fmt.Errorf("failed to read migration directory: %w", err)
    }

    upFiles := make([]string, 0)
    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".up.sql") {
            upFiles = append(upFiles, file.Name())
        }
    }

    if len(upFiles) == 0 {
        fmt.Println("No migration files found.")
        return nil
    }

    fmt.Printf("Available migrations (%d):\n", len(upFiles))
    for i, file := range upFiles {
        // Extract timestamp and name
        parts := strings.SplitN(file, "_", 2)
        if len(parts) >= 2 {
            timestamp := parts[0]
            name := strings.TrimSuffix(parts[1], ".up.sql")
            fmt.Printf("   %d. %s - %s\n", i+1, timestamp, name)
        } else {
            fmt.Printf("   %d. %s\n", i+1, file)
        }
    }

    return nil
}

// GetMigrationStatus shows current migration status with details
func (m *MigrationRunner) GetMigrationStatus() error {
    version, dirty, err := m.GetVersion()
    if err != nil {
        return err
    }

    fmt.Printf("Migration Status:\n")
    fmt.Printf("   Current Version: %d\n", version)
    fmt.Printf("   Dirty State: %t\n", dirty)
    
    if dirty {
        fmt.Printf("   Warning: Database is in a dirty state!\n")
        fmt.Printf("   This usually means a migration failed partway through.\n")
        fmt.Printf("   You may need to manually fix the database or use 'make migrate-force'\n")
    } else {
        fmt.Printf("   Database is clean\n")
    }

    return nil
}


// getMigrator creates and returns a new migrator instance.
func (m *MigrationRunner) getMigrator() (*migrate.Migrate, error) {
	sqlDB := stdlib.OpenDB(*m.pool.Config().ConnConfig)
	
	// This assumes migrations.Files is the embed.FS from another package
	sourceDriver, err := iofs.New(migrations.Files, ".") 
	if err != nil {
		return nil, fmt.Errorf("failed to create migration source: %w", err)
	}

	databaseDriver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create database driver: %w", err)
	}

	migrator, err := migrate.NewWithInstance(
		"iofs",         // sourceName
		sourceDriver,   // source.Driver
		"postgres",     // databaseName
		databaseDriver, // database.Driver
	)
	if err != nil {
		// Close sqlDB if migrator creation fails
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}
	// The migrator.Close() will handle closing the underlying sqlDB
	return migrator, nil
}

func (m *MigrationRunner) RunMigrations() error {
	migrator, err := m.getMigrator()
	if err != nil {
		return err
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// RunDownMigration rolls back the last applied migration.
func (m *MigrationRunner) RunDownMigration() error {
	migrator, err := m.getMigrator()
	if err != nil {
		return err
	}
	defer migrator.Close()

	if err := migrator.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to roll back last migration: %w", err)
	}
	return nil
}

// GetVersion returns the current database schema version and if it's dirty.
func (m *MigrationRunner) GetVersion() (uint, bool, error) {
	migrator, err := m.getMigrator()
	if err != nil {
		return 0, false, err
	}
	defer migrator.Close()

	version, dirty, err := migrator.Version()
	if err != nil && err != migrate.ErrNoChange {
		return 0, false, fmt.Errorf("failed to get database version: %w", err)
	}
	return version, dirty, nil
}

// ForceVersion forces the database version to a specific number. Use with caution.
func (m *MigrationRunner) ForceVersion(version int) error {
	migrator, err := m.getMigrator()
	if err != nil {
		return err
	}
	defer migrator.Close()

	if err := migrator.Force(version); err != nil {
		return fmt.Errorf("failed to force database version: %w", err)
	}
	return nil
}
