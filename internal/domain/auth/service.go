// domain/auth/service.go
package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/shared/types"
	"github.com/purushothdl/gochat-backend/pkg/auth"
)

const MaxDevicesPerUser = 5

type Service struct {
    authRepo Repository
    userRepo UserRepository
    config   *config.Config
    logger   *slog.Logger
}

func NewService(authRepo Repository, userRepo UserRepository, cfg *config.Config, logger *slog.Logger) *Service {
    return &Service{
        authRepo: authRepo,
        userRepo: userRepo,
        config:   cfg,
        logger:   logger,
    }
}

// ============================================================================
//  Public Methods (The Service's API)
// ============================================================================

func (s *Service) Login(ctx context.Context, req LoginRequest, w http.ResponseWriter, deviceInfo, ipAddress, userAgent string) (*AuthenticationResponse, error) {
    // Get user 
    user, err := s.userRepo.GetByEmailShared(ctx, req.Email)
    if err != nil {
        return nil, ErrInvalidCredentials
    }
    
    // Get password hash and verify
    hashedPassword, err := s.userRepo.GetPasswordHash(ctx, user.ID)
    if err != nil {
        return nil, ErrInvalidCredentials
    }
    
    if !auth.CheckPassword(req.Password, hashedPassword) {
        return nil, ErrInvalidCredentials
    }
    
    // Update last login
    s.userRepo.UpdateLastLogin(ctx, user.ID)
    
    // Generate tokens and set cookie
    return s.generateAuthResponse(ctx, user, w, deviceInfo, ipAddress, userAgent)
}

func (s *Service) Register(ctx context.Context, req RegisterRequest, w http.ResponseWriter, deviceInfo, ipAddress, userAgent string) (*AuthenticationResponse, error) {
    // Check if user exists
    exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, ErrUserAlreadyExists
    }
    
    // Hash password
    hashedPassword, err := auth.HashPassword(req.Password, s.config.Security.BcryptCost)
    if err != nil {
        return nil, err
    }
    
    // Create user
    userData := &types.CreateUserData{
        Email:    req.Email,
        Name:     req.Name,
        Password: hashedPassword,
    }
    
    user, err := s.userRepo.Create(ctx, userData)
    if err != nil {
        return nil, err
    }
    
    return s.generateAuthResponse(ctx, user, w, deviceInfo, ipAddress, userAgent)
}

func (s *Service) RefreshAccessToken(ctx context.Context, refreshTokenString, deviceInfo, ipAddress, userAgent string) (*RefreshTokenResponse, error) {
    tokenHash := auth.HashRefreshToken(refreshTokenString)
    
    // Get refresh token from database
    refreshToken, err := s.authRepo.GetRefreshToken(ctx, tokenHash)
    if err != nil {
        return nil, ErrRefreshTokenNotFound
    }
    
    // Check if token is expired 
    if refreshToken.ExpiresAt.Before(time.Now()) {
        s.authRepo.DeleteRefreshToken(ctx, tokenHash)
        return nil, ErrTokenExpired
    }
    
    // Get user info to include email in JWT
    user, err := s.userRepo.GetByIDShared(ctx, refreshToken.UserID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    // Update token usage
    if err := s.authRepo.UpdateRefreshTokenUsage(ctx, tokenHash); err != nil {
        // Log error but continue
    }
    
    // Generate new access token with user email
    accessToken, err := auth.GenerateAccessToken(
        user.ID,
        user.Email,
        s.config.JWT.Secret,
        s.config.JWT.AccessTokenExpiry,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }
    
    return &RefreshTokenResponse{
        AccessToken: accessToken,
        ExpiresAt:   time.Now().Add(s.config.JWT.AccessTokenExpiry),
    }, nil
}

func (s *Service) GetMe(ctx context.Context, userID string) (*MeResponse, error) {
    user, err := s.userRepo.GetByIDShared(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return &MeResponse{
        User: UserInfo{
            ID:         user.ID,
            Email:      user.Email,
            Name:       user.Name,
            ImageURL:   user.ImageURL,
            IsVerified: user.IsVerified,
            CreatedAt:  user.CreatedAt,
        },
    }, nil
}

func (s *Service) Logout(ctx context.Context, refreshTokenString string) error {
    if refreshTokenString == "" {
        return nil // Already logged out
    }
    
    tokenHash := auth.HashRefreshToken(refreshTokenString)
    return s.authRepo.DeleteRefreshToken(ctx, tokenHash)
}

func (s *Service) LogoutAllDevices(ctx context.Context, userID string) error {
    return s.authRepo.DeleteUserRefreshTokens(ctx, userID)
}

// ============================================================================
//  Private Helper Methods (Implementation Details)
// ============================================================================

func (s *Service) generateAuthResponse(ctx context.Context, user *types.User, w http.ResponseWriter, deviceInfo, ipAddress, userAgent string) (*AuthenticationResponse, error) {
    // Generate access token
    // Inside generateAuthResponse
    s.logger.Info("Generating access token", "userID", user.ID, "email", user.Email)

    accessToken, err := auth.GenerateAccessToken(
        user.ID, user.Email, s.config.JWT.Secret, s.config.JWT.AccessTokenExpiry,
    )
    if err != nil {
        return nil, err
    }
    
    // Generate refresh token
    refreshTokenString, err := auth.GenerateRefreshToken()
    if err != nil {
        return nil, err
    }
    
    // Enforce device limit
    s.enforceDeviceLimit(ctx, user.ID)
    
    // Create refresh token record
    refreshToken := &RefreshToken{
        ID:         uuid.New().String(),
        UserID:     user.ID,
        TokenHash:  auth.HashRefreshToken(refreshTokenString),
        DeviceInfo: deviceInfo,
        IPAddress:  ipAddress,
        UserAgent:  userAgent,
        ExpiresAt:  time.Now().Add(s.config.JWT.RefreshTokenExpiry),
    }
    
    if err := s.authRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
        return nil, err
    }
    
    // Set HTTP-only cookie with refresh token
    s.setRefreshTokenCookie(w, refreshTokenString)
    
    return &AuthenticationResponse{
        AccessToken: accessToken,
        ExpiresAt:   time.Now().Add(s.config.JWT.AccessTokenExpiry),
        User: UserInfo{
            ID:         user.ID,
            Email:      user.Email,
            Name:       user.Name,
            ImageURL:   user.ImageURL,
            IsVerified: user.IsVerified,
            CreatedAt:  user.CreatedAt,
        },
    }, nil
}

// Set secure HTTP-only cookie
func (s *Service) setRefreshTokenCookie(w http.ResponseWriter, refreshToken string) {
    cookie := &http.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken,
        Path:     "/",
        MaxAge:   int(s.config.JWT.RefreshTokenExpiry.Seconds()),
        HttpOnly: true,
        Secure:   s.config.Server.Env == "production",
        SameSite: http.SameSiteStrictMode,
    }
    
    http.SetCookie(w, cookie)
}

func (s *Service) enforceDeviceLimit(ctx context.Context, userID string) error {
    count, err := s.authRepo.CountUserTokens(ctx, userID)
    if err != nil {
        return err
    }
    
    if count >= MaxDevicesPerUser {
        return s.authRepo.DeleteOldestUserToken(ctx, userID)
    }
    
    return nil
}