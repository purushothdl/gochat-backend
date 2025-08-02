//internal/transport/http/middleware/recovery.go`:
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/purushothdl/gochat-backend/internal/shared/response"
	"github.com/purushothdl/gochat-backend/pkg/errors"
)

// Recoverer prevents panics from crashing the server and logs them gracefully.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						"error", err,
						"request_method", r.Method,
						"request_path", r.URL.Path,
						"stack", string(debug.Stack()),
					)

					response.Error(w, http.StatusInternalServerError, errors.ErrInternalServer)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}