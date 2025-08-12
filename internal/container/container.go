package container

import (
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/contracts" 
	"github.com/purushothdl/gochat-backend/internal/domain/auth"
	"github.com/purushothdl/gochat-backend/internal/domain/health"
	"github.com/purushothdl/gochat-backend/internal/domain/message"
	"github.com/purushothdl/gochat-backend/internal/domain/room"
	"github.com/purushothdl/gochat-backend/internal/domain/upload"
	"github.com/purushothdl/gochat-backend/internal/domain/user"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/email"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/imageproc"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/postgres"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/redis"
	"github.com/purushothdl/gochat-backend/internal/infrastructure/s3"
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
	MessageRepo       *postgres.MessageRepository
	
	// Infrastructure Providers (implementing contracts)
	QueueProvider   contracts.Queue
	StorageProvider contracts.FileStorage
	EmailService    *email.ResendService
	ImageProcessor  *imageproc.Processor

	// Domain Services
	AuthService    *auth.Service
	UserService    *user.Service
	HealthService  *health.Service
	RoomService    *room.Service
	MessageService *message.Service
	UploadService  *upload.Service

	// Workers
	UploadWorker *upload.Worker

	// Handlers
	AuthHandler    *auth.Handler
	UserHandler    *user.Handler
	HealthHandler  *health.Handler
	RoomHandler    *room.Handler
	MessageHandler *message.Handler

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
	var err error

	// Build Repositories
	c.UserRepo = postgres.NewUserRepository(c.DB)
	c.AuthRepo = postgres.NewAuthRepository(c.DB)
	c.PasswordResetRepo = postgres.NewPasswordResetRepository(c.DB)
	c.RoomRepo = postgres.NewRoomRepository(c.DB)
	c.MessageRepo = postgres.NewMessageRepository(c.DB)

	// Build Infrastructure Providers
	c.StorageProvider, err = s3.NewClient(&c.Config.AWS)
	if err != nil {
		return fmt.Errorf("failed to create s3 client: %w", err)
	}
	c.QueueProvider, err = redis.NewQueueProvider(&c.Config.Redis)
	if err != nil {
		return fmt.Errorf("failed to create queue provider: %w", err)
	}
	c.EmailService = email.NewResendService(&c.Config.Resend)
	c.ImageProcessor = imageproc.NewProcessor()

	// Build Domain Services
	c.AuthService = auth.NewService(c.AuthRepo, c.UserRepo, c.PasswordResetRepo, c.EmailService, c.Config, c.Logger)
	c.UserService = user.NewService(c.UserRepo, c.Config, c.Logger)
	c.HealthService = health.NewService(c.DB, c.Logger)
	c.RoomService = room.NewService(c.RoomRepo, c.UserRepo, c.Config, c.Logger)
	c.MessageService = message.NewService(c.MessageRepo, c.RoomRepo, c.UserRepo, nil, c.Config, c.Logger)
	
	// The upload.Service fulfills the user.ProfileImageUploader interface implicitly.
	c.UploadService = upload.NewService(c.StorageProvider, c.QueueProvider, c.ImageProcessor, c.Config, c.Logger)

	// Build Workers
	c.UploadWorker = upload.NewWorker(c.QueueProvider, c.StorageProvider, c.UserRepo, c.ImageProcessor, c.Config, c.Logger)

	// Build Handlers
	c.AuthHandler = auth.NewHandler(c.AuthService, c.Logger, c.Validator)
	c.UserHandler = user.NewHandler(c.UserService, c.Logger, c.Validator, c.Config, c.UploadService) 
	c.HealthHandler = health.NewHandler(c.HealthService, c.Logger)
	c.RoomHandler = room.NewHandler(c.RoomService, c.Logger, c.Validator)
	c.MessageHandler = message.NewHandler(c.MessageService, c.Logger, c.Validator)

	// Build Middleware
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.Config, c.UserRepo)

	return nil
}