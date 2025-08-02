package container

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/domain/auth"
	"github.com/purushothdl/gochat-backend/internal/domain/user"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/email"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/postgres"
	"github.com/purushothdl/gochat-backend/internal/transport/http/middleware"
)

// Container holds all application dependencies.
type Container struct {
	Config *config.Config
	DB     *pgxpool.Pool
	Logger *slog.Logger

	// Repositories
	UserRepo          *postgres.UserRepository
	AuthRepo          *postgres.AuthRepository
	PasswordResetRepo *postgres.PasswordResetRepository

	// External Services
	EmailService *email.ResendService

	// Domain Services
	AuthService *auth.Service
	UserService *user.Service

	// Handlers
	AuthHandler *auth.Handler
	UserHandler *user.Handler

	// Middleware
	AuthMiddleware *middleware.AuthMiddleware
}

// New creates a new, uninitialized container.
func New(cfg *config.Config, db *pgxpool.Pool, logger *slog.Logger) *Container {
	return &Container{
		Config: cfg,
		DB:     db,
		Logger: logger,
	}
}

// Build constructs and wires all dependencies.
func (c *Container) Build() error {
	// Build repositories
	c.UserRepo = postgres.NewUserRepository(c.DB)
	c.AuthRepo = postgres.NewAuthRepository(c.DB)
	c.PasswordResetRepo = postgres.NewPasswordResetRepository(c.DB)

	// Build external service clients
	c.EmailService = email.NewResendService(&c.Config.Resend)

	// Build domain services, injecting their dependencies
	c.AuthService = auth.NewService(
		c.AuthRepo,
		c.UserRepo,
		c.PasswordResetRepo,
		c.EmailService,
		c.Config,
		c.Logger,
	)
	c.UserService = user.NewService(c.UserRepo, c.Config, c.Logger)

	// Build handlers
	c.AuthHandler = auth.NewHandler(c.AuthService, c.Logger)
	c.UserHandler = user.NewHandler(c.UserService, c.Logger)

	// Build middleware
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.Config)

	return nil
}