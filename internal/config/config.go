package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Security SecurityConfig
}

type ServerConfig struct {
    Port string
    Env  string
}

type DatabaseConfig struct {
    URL             string
    Host            string
    Port            string
    User            string
    Password        string
    Name            string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

type JWTConfig struct {
    Secret             string
    AccessTokenExpiry  time.Duration
    RefreshTokenExpiry time.Duration
}

type SecurityConfig struct {
    BcryptCost int
}

func Load() (*Config, error) {
    return &Config{
        Server: ServerConfig{
            Port: getEnv("PORT", "8080"),
            Env:  getEnv("ENV", "development"),
        },
        Database: DatabaseConfig{
            URL:             getEnv("DATABASE_URL", ""),
            Host:            getEnv("DB_HOST", "localhost"),
            Port:            getEnv("DB_PORT", "5432"),
            User:            getEnv("DB_USER", ""),
            Password:        getEnv("DB_PASSWORD", ""),
            Name:            getEnv("DB_NAME", "gochat"),
            MaxOpenConns:    parseInt("DB_MAX_OPEN_CONNS", 25),
            MaxIdleConns:    parseInt("DB_MAX_IDLE_CONNS", 5),
            ConnMaxLifetime: parseDuration("DB_CONN_MAX_LIFETIME", "5m"),
            ConnMaxIdleTime: parseDuration("DB_CONN_MAX_IDLE_TIME", "5m"),
        },
        JWT: JWTConfig{
            Secret:             getEnv("JWT_SECRET", "your-secret-key"),
            AccessTokenExpiry:  parseDuration("JWT_EXPIRY", "15m"),
            RefreshTokenExpiry: parseDuration("REFRESH_TOKEN_EXPIRY", "30d"),
        },
        Security: SecurityConfig{
            BcryptCost: parseInt("BCRYPT_COST", 12),
        },
    }, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func parseDuration(key, defaultValue string) time.Duration {
    if value := os.Getenv(key); value != "" {
        if d, err := time.ParseDuration(value); err == nil {
            return d
        }
    }
    d, _ := time.ParseDuration(defaultValue)
    return d
}

func parseInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if i, err := strconv.Atoi(value); err == nil {
            return i
        }
    }
    return defaultValue
}
