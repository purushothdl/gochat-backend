package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Security SecurityConfig
	CORS     CORSConfig
	Resend   ResendConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	ClientURL string
}

type ServerConfig struct {
	Port            string
	Env             string
	ReadTimeout     time.Duration // Time for reading the entire request
	WriteTimeout    time.Duration // Time for writing the entire response
	IdleTimeout     time.Duration // Time for a keep-alive connection to be idle
	ShutdownTimeout time.Duration // Max time for graceful shutdown
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

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type ResendConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
}

func Load() (*Config, error) {
    return &Config{
        App: AppConfig{
			ClientURL: getEnv("CLIENT_URL", "http://localhost:3000"),
		},
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			Env:             getEnv("ENV", "development"),
			ReadTimeout:     parseDuration("SERVER_READ_TIMEOUT", "5s"),
			WriteTimeout:    parseDuration("SERVER_WRITE_TIMEOUT", "10s"),
			IdleTimeout:     parseDuration("SERVER_IDLE_TIMEOUT", "120s"),
			ShutdownTimeout: parseDuration("SERVER_SHUTDOWN_TIMEOUT", "30s"),
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
        CORS: CORSConfig{
			AllowedOrigins:   strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"), ","),
			AllowedMethods:   strings.Split(getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"), ","),
			AllowedHeaders:   strings.Split(getEnv("CORS_ALLOWED_HEADERS", "Accept,Authorization,Content-Type"), ","),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		},
        Resend: ResendConfig{
			APIKey:    getEnv("RESEND_API_KEY", ""),
			FromEmail: getEnv("RESEND_FROM_EMAIL", "onboarding@resend.dev"),
			FromName:  getEnv("RESEND_FROM_NAME", "GoChat"),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}