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

type authContextKey struct {
    name string
}

var (
    userIDKey   = &authContextKey{"userID"}
    userEmailKey = &authContextKey{"userEmail"}
)

type AuthMiddleware struct {
    config *config.Config
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
    return &AuthMiddleware{config: cfg}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
            return
        }

        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
            return
        }

        claims, err := auth.ValidateAccessToken(tokenParts[1], m.config.JWT.Secret)
        if err != nil {
            response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
        ctx = context.WithValue(ctx, userEmailKey, claims.Email)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")

        if authHeader != "" {
            tokenParts := strings.Split(authHeader, " ")
            if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
                if claims, err := auth.ValidateAccessToken(tokenParts[1], m.config.JWT.Secret); err == nil {
                    ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
                    ctx = context.WithValue(ctx, userEmailKey, claims.Email)
                    r = r.WithContext(ctx)
                }
            }
        }

        next.ServeHTTP(w, r)
    })
}

func GetUserID(ctx context.Context) (string, bool) {
    userID, ok := ctx.Value(userIDKey).(string)
    return userID, ok
}

func GetUserEmail(ctx context.Context) (string, bool) {
    email, ok := ctx.Value(userEmailKey).(string)
    return email, ok
}