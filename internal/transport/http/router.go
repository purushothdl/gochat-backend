package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/domain/auth"
	"github.com/purushothdl/gochat-backend/internal/domain/user"
	app_middleware "github.com/purushothdl/gochat-backend/internal/transport/http/middleware"
)

type Router struct {
	authHandler *auth.Handler
	userHandler *user.Handler
	authMw      *app_middleware.AuthMiddleware
}

func NewRouter(authHandler *auth.Handler, userHandler *user.Handler, authMw *app_middleware.AuthMiddleware) *Router {
	return &Router{
		authHandler: authHandler,
		userHandler: userHandler,
		authMw:      authMw,
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

	// Health check route
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", rt.authHandler.Register)
			r.Post("/login", rt.authHandler.Login)
			r.Post("/refresh", rt.authHandler.RefreshToken)
			r.Post("/logout", rt.authHandler.Logout)
            r.Post("/forgot-password", rt.authHandler.ForgotPassword)
            r.Post("/reset-password", rt.authHandler.ResetPassword) 
			r.With(rt.authMw.RequireAuth).Get("/me", rt.authHandler.Me)
		})

		r.Route("/user", func(r chi.Router) {
			r.Use(rt.authMw.RequireAuth)
			r.Get("/profile", rt.userHandler.GetProfile)
			r.Put("/profile", rt.userHandler.UpdateProfile)
			r.Put("/settings", rt.userHandler.UpdateSettings)
			r.Put("/password", rt.userHandler.ChangePassword)
		})
	})

	return r
}