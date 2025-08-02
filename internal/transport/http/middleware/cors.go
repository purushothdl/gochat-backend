package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
	"github.com/purushothdl/gochat-backend/internal/config"
)

// CORS configures and returns a new CORS handler from settings.
func CORS(cfg *config.CORSConfig) func(http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           300,
	}).Handler
}