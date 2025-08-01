package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/purushothdl/gochat-backend/internal/config"
    "github.com/purushothdl/gochat-backend/internal/shared/response"
    "github.com/purushothdl/gochat-backend/pkg/auth"
    "github.com/purushothdl/gochat-backend/pkg/errors"
)

type AuthMiddleware struct {
    config *config.Config
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
    return &AuthMiddleware{config: cfg}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
            return
        }

        // Check Bearer token format
        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
            return
        }

        // Validate JWT token
        claims, err := auth.ValidateAccessToken(tokenParts[1], m.config.JWT.Secret)
        if err != nil {
            response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
            return
        }

        // Add user info to context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "user_email", claims.Email)

        // Call next handler
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}

// Optional middleware for routes that can work with or without auth
func (m *AuthMiddleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        
        if authHeader != "" {
            tokenParts := strings.Split(authHeader, " ")
            if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
                if claims, err := auth.ValidateAccessToken(tokenParts[1], m.config.JWT.Secret); err == nil {
                    ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
                    ctx = context.WithValue(ctx, "user_email", claims.Email)
                    r = r.WithContext(ctx)
                }
            }
        }

        next.ServeHTTP(w, r)
    }
}
