package http

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/domain/auth"
	"github.com/purushothdl/gochat-backend/internal/domain/health"
	"github.com/purushothdl/gochat-backend/internal/domain/room"
	"github.com/purushothdl/gochat-backend/internal/domain/user"
	app_middleware "github.com/purushothdl/gochat-backend/internal/transport/http/middleware"
)

type Router struct {
	authHandler   *auth.Handler
	userHandler   *user.Handler
	healthHandler *health.Handler
	roomHandler   *room.Handler
	authMw        *app_middleware.AuthMiddleware
}

func NewRouter(authHandler *auth.Handler, userHandler *user.Handler, healthHandler *health.Handler, roomHandler *room.Handler, authMw *app_middleware.AuthMiddleware) *Router {
	return &Router{
		authHandler:   authHandler,
		userHandler:   userHandler,
		healthHandler: healthHandler,
		roomHandler:   roomHandler,
		authMw:        authMw,
	}
}

// The order is important: recovery -> cors -> logger -> request logger -> timeout.
func (rt *Router) mountMiddlewares(r *chi.Mux, cfg *config.Config, logger *slog.Logger) {
	r.Use(app_middleware.Recoverer(logger))
	r.Use(app_middleware.CORS(&cfg.CORS))
	r.Use(app_middleware.WithLogger(logger))
	r.Use(app_middleware.RequestLogger)
	r.Use(middleware.Timeout(60 * time.Second))
}

func (rt *Router) SetupRoutes(cfg *config.Config, logger *slog.Logger) *chi.Mux {
	r := chi.NewMux()

	rt.mountMiddlewares(r, cfg, logger)

	// Health check routes (outside /api for load balancers)
	r.Get("/health", rt.healthHandler.Health)
	r.Get("/ready", rt.healthHandler.Ready)
	r.Get("/live", rt.healthHandler.Live)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", rt.authHandler.Register)               // Register a new user
			r.Post("/login", rt.authHandler.Login)                     // User login
			r.Post("/refresh", rt.authHandler.RefreshToken)            // Refresh access token
			r.Post("/logout", rt.authHandler.Logout)                   // User logout
			r.Post("/forgot-password", rt.authHandler.ForgotPassword)  // Initiate password reset
			r.Post("/reset-password", rt.authHandler.ResetPassword)    // Complete password reset
			r.With(rt.authMw.RequireAuth).Get("/me", rt.authHandler.Me) // Get current user info
		})

		r.Route("/user", func(r chi.Router) {
			r.Use(rt.authMw.RequireAuth)
			r.Get("/profile", rt.userHandler.GetProfile)               // Get user profile
			r.Put("/profile", rt.userHandler.UpdateProfile)            // Update user profile
			r.Put("/settings", rt.userHandler.UpdateSettings)          // Update user settings
			r.Put("/password", rt.userHandler.ChangePassword)          // Change user password
		})

		r.Route("/rooms", func(r chi.Router) {
			r.Use(rt.authMw.RequireAuth)
			r.Post("/", rt.roomHandler.CreateRoom)                     // Create a new room
			r.Post("/{room_id}/invite", rt.roomHandler.InviteUser)     // Invite user to a room
			r.Get("/", rt.roomHandler.ListUserRooms)                   // List rooms for the authenticated user
			r.Get("/public", rt.roomHandler.ListPublicRooms)           // List all public rooms
			r.Get("/{room_id}/members", rt.roomHandler.ListMembers)    // List members of a specific room
			r.Post("/{room_id}/join", rt.roomHandler.JoinPublicRoom)   // Join a public room
		})
	})

	return r
}
