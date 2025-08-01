package container

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/domain/auth"
	"github.com/purushothdl/gochat-backend/internal/domain/user"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/postgres"
	"github.com/purushothdl/gochat-backend/internal/transport/http/middleware"
)

type Container struct {
    Config *config.Config
    DB     *pgxpool.Pool

    // Repositories
    UserRepo *postgres.UserRepository
    AuthRepo *postgres.AuthRepository

    // Services
    AuthService *auth.Service
    UserService *user.Service

    // Handlers
    AuthHandler *auth.Handler
    UserHandler *user.Handler

    // Middleware
    AuthMiddleware *middleware.AuthMiddleware
}

func New(cfg *config.Config, db *pgxpool.Pool) *Container {
    return &Container{
        Config: cfg,
        DB:     db,
    }
}

func (c *Container) Build() error {
    // Build repositories
    c.UserRepo = postgres.NewUserRepository(c.DB)
    c.AuthRepo = postgres.NewAuthRepository(c.DB)

    // Build services
    c.AuthService = auth.NewService(c.AuthRepo, c.UserRepo, c.Config)
    c.UserService = user.NewService(c.UserRepo, c.Config)

    // Build handlers
    c.AuthHandler = auth.NewHandler(c.AuthService)
    c.UserHandler = user.NewHandler(c.UserService)

    // Build middleware
    c.AuthMiddleware = middleware.NewAuthMiddleware(c.Config)

    return nil
}
