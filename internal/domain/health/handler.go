package health

import (
    "context"
    "encoding/json"
    "log/slog"
    "net/http"
    "time"
)

type Handler struct {
    service *Service
    logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

// Health returns detailed health information
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    health := h.service.CheckHealth(ctx)
    
    statusCode := http.StatusOK
    if health.Status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    } else if health.Status == "degraded" {
        statusCode = http.StatusPartialContent
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    if err := json.NewEncoder(w).Encode(health); err != nil {
        h.logger.Error("Failed to encode health response", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
}

// Ready returns a simple readiness check (for load balancers)
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    if h.service.IsHealthy(ctx) {
        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    } else {
        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("Not Ready"))
    }
}

// Live returns a simple liveness check
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
    // For now, just return OK if the handler can respond
    // Can add more sophisticated checks later, if needed
    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
