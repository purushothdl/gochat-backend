package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/purushothdl/gochat-backend/internal/config"
)

// Server encapsulates the HTTP server, its configuration, and dependencies.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	config     *config.Config
}

// NewServer creates and configures a new Server instance.
func NewServer(cfg *config.Config, logger *slog.Logger, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
			ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		},
		logger: logger,
		config: cfg,
	}
}

// Run starts the server and blocks until a shutdown signal is received.
func (s *Server) Run() error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 1)
	go func() {
		s.logger.Info("server starting", "address", s.httpServer.Addr)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		s.logger.Info("shutdown signal received", "signal", sig)
		s.gracefulShutdown()
	}

	return nil
}

// gracefulShutdown attempts to shut down the server gracefully.
func (s *Server) gracefulShutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("graceful shutdown failed", "error", err)
		if err := s.httpServer.Close(); err != nil {
			s.logger.Error("could not close server", "error", err)
		}
	} else {
		s.logger.Info("server shutdown complete")
	}
}