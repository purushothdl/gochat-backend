package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/shared/response"
	"github.com/purushothdl/gochat-backend/internal/shared/types"
	"github.com/purushothdl/gochat-backend/pkg/auth"
	"github.com/purushothdl/gochat-backend/pkg/errors"
)

type authContextKey struct {
	name string
}

var (
	userIDKey    = &authContextKey{"userID"}
	userEmailKey = &authContextKey{"userEmail"}
	userKey      = &authContextKey{"user"}
	deviceIDKey  = &authContextKey{"deviceID"}
)

// Required by auth middleware to set the current user
type UserRepository interface {
	GetByIDShared(ctx context.Context, userID string) (*types.User, error)
}

type AuthMiddleware struct {
	config   *config.Config
	userRepo UserRepository
}

func NewAuthMiddleware(cfg *config.Config, userRepo UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		config:   cfg,
		userRepo: userRepo,
	}
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
		ctx = context.WithValue(ctx, deviceIDKey, claims.DeviceID)

		// Fetch the full User entity from the repository
		userEntity, err := m.userRepo.GetByIDShared(ctx, claims.UserID)
		if err != nil {
			// It's possible the user doesn't exist or a DB error occurred
			response.Error(w, http.StatusInternalServerError, errors.ErrInternalServer)
			return
		}

		// Convert the full User entity to BasicUser before storing in context
		basicUser := &types.BasicUser{
			ID:       userEntity.ID,
			Name:     userEntity.Name,
			ImageURL: userEntity.ImageURL,
		}
		ctx = context.WithValue(ctx, userKey, basicUser)

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

					// Fetch the full User entity from the repository
					userEntity, err := m.userRepo.GetByIDShared(ctx, claims.UserID)
					if err == nil {
						basicUser := &types.BasicUser{
							ID:       userEntity.ID,
							Name:     userEntity.Name,
							ImageURL: userEntity.ImageURL,
						}
						ctx = context.WithValue(ctx, userKey, basicUser)
					}
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

func GetBasicUser(ctx context.Context) (*types.BasicUser, bool) {
	basicUser, ok := ctx.Value(userKey).(*types.BasicUser)
	return basicUser, ok
}

func GetDeviceID(ctx context.Context) (string, bool) {
	deviceID, ok := ctx.Value(deviceIDKey).(string)
	return deviceID, ok
}
