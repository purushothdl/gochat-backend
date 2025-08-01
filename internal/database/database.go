package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/purushothdl/gochat-backend/internal/config"
)

func Connect(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
    var dsn string
    if cfg.URL != "" {
        dsn = cfg.URL
    } else {
        dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
            cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
    }

    // Use pgxpool.New to create a connection pool
    pool, err := pgxpool.New(context.Background(), dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }

    // Configure connection pool
    pool.Config().MaxConns = int32(cfg.MaxOpenConns)
    pool.Config().MinConns = int32(cfg.MaxIdleConns)
    pool.Config().MaxConnLifetime = cfg.ConnMaxLifetime
    pool.Config().MaxConnIdleTime = cfg.ConnMaxIdleTime

    // Ping the database
    if err := pool.Ping(context.Background()); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return pool, nil
}
