package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/postgres/migrations"
)

type MigrationRunner struct {
	pool *pgxpool.Pool
}

func NewMigrationRunner(pool *pgxpool.Pool) *MigrationRunner {
	return &MigrationRunner{pool: pool}
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
