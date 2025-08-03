package container

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/domain/auth"
	"github.com/purushothdl/gochat-backend/internal/domain/health"
	"github.com/purushothdl/gochat-backend/internal/domain/room" 
	"github.com/purushothdl/gochat-backend/internal/domain/user"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/email"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/postgres"
	"github.com/purushothdl/gochat-backend/internal/shared/validator"
	"github.com/purushothdl/gochat-backend/internal/transport/http/middleware"
)

type Container struct {
	Config    *config.Config
	DB        *pgxpool.Pool
	Logger    *slog.Logger
	Validator *validator.Validator

	// Repositories
	UserRepo          *postgres.UserRepository
	AuthRepo          *postgres.AuthRepository
	PasswordResetRepo *postgres.PasswordResetRepository
	RoomRepo          *postgres.RoomRepository 
	// External Services
	EmailService *email.ResendService

	// Domain Services
	AuthService   *auth.Service
	UserService   *user.Service
	HealthService *health.Service
	RoomService   *room.Service 

	// Handlers
	AuthHandler   *auth.Handler
	UserHandler   *user.Handler
	HealthHandler *health.Handler
	RoomHandler   *room.Handler 

	// Middleware
	AuthMiddleware *middleware.AuthMiddleware
}

func New(cfg *config.Config, db *pgxpool.Pool, logger *slog.Logger) *Container {
	return &Container{
		Config:    cfg,
		DB:        db,
		Logger:    logger,
		Validator: validator.New(),
	}
}

func (c *Container) Build() error {
	// Build repositories
	c.UserRepo = postgres.NewUserRepository(c.DB)
	c.AuthRepo = postgres.NewAuthRepository(c.DB)
	c.PasswordResetRepo = postgres.NewPasswordResetRepository(c.DB)
	c.RoomRepo = postgres.NewRoomRepository(c.DB) 

	// Build external service clients
	c.EmailService = email.NewResendService(&c.Config.Resend)

	// Build domain services
	c.AuthService = auth.NewService(
		c.AuthRepo,
		c.UserRepo,
		c.PasswordResetRepo,
		c.EmailService,
		c.Config,
		c.Logger,
	)
	c.UserService = user.NewService(c.UserRepo, c.Config, c.Logger)
	c.HealthService = health.NewService(c.DB, c.Logger)
	c.RoomService = room.NewService(c.RoomRepo, c.UserRepo, c.Config, c.Logger) 

	// Build handlers
	c.AuthHandler = auth.NewHandler(c.AuthService, c.Logger, c.Validator)
	c.UserHandler = user.NewHandler(c.UserService, c.Logger, c.Validator)
	c.HealthHandler = health.NewHandler(c.HealthService, c.Logger)
	c.RoomHandler = room.NewHandler(c.RoomService, c.Logger, c.Validator) 

	// Build middleware
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.Config)

	return nil
}
