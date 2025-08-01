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

func (m *MigrationRunner) RunMigrations() error {
	// Convert pgxpool to sql.DB for migrate
	sqlDB := stdlib.OpenDB(*m.pool.Config().ConnConfig)
	defer sqlDB.Close()

	// Create migration source from embedded files
	sourceDriver, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create database driver
	databaseDriver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	// Create migrator
	migrator, err := migrate.NewWithInstance(
		"iofs",         // sourceName
		sourceDriver,   // source.Driver
		"postgres",     // databaseName
		databaseDriver, // database.Driver
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	// Run migrations
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
