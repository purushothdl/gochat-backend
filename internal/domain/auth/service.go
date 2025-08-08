// domain/auth/service.go
package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/shared/types"
	"github.com/purushothdl/gochat-backend/pkg/auth"
	"github.com/purushothdl/gochat-backend/pkg/utils/tokenutil"
)

const MaxDevicesPerUser = 5

type Service struct {
	authRepo          Repository
	userRepo          UserRepository
	config            *config.Config
	logger            *slog.Logger
	passwordResetRepo PasswordResetRepository
	emailService      EmailService
}

func NewService(
	authRepo Repository,
	userRepo UserRepository,
	passwordResetRepo PasswordResetRepository,
	emailService EmailService,
	cfg *config.Config,
	logger *slog.Logger,
) *Service {
	return &Service{
		authRepo:          authRepo,
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		emailService:      emailService,
		config:            cfg,
		logger:            logger,
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

// ForgotPassword generates a reset token and sends it via email.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmailShared(ctx, email)
	if err != nil {
		// Do not leak whether the user exists. Log internally and return success.
		s.logger.Info("password reset requested for non-existent user", "email", email)
		return nil
	}

	plaintext, hash, err := tokenutil.Generate()
	if err != nil {
		s.logger.Error("failed to generate password reset token", "error", err)
		return err
	}

	// Store the hashed token with a 1-hour expiry.
	expiresAt := time.Now().Add(1 * time.Hour)
	if err := s.passwordResetRepo.Store(ctx, user.ID, hash, expiresAt); err != nil {
		s.logger.Error("failed to store password reset token", "user_id", user.ID, "error", err)
		return err
	}

	// Construct the reset link.
	resetURL, err := url.Parse(s.config.App.ClientURL)
	if err != nil {
		s.logger.Error("invalid client url in config", "url", s.config.App.ClientURL, "error", err)
		return err
	}
	resetURL.Path = "/reset-password"
	query := resetURL.Query()
	query.Set("token", plaintext)
	resetURL.RawQuery = query.Encode()

	// Send the email in a separate goroutine to not block the HTTP response.
	go func() {
		err := s.emailService.SendPasswordResetEmail(context.Background(), user.Email, user.Name, resetURL.String())
		if err != nil {
			s.logger.Error("failed to send password reset email", "user_id", user.ID, "error", err)
		}
	}()

	return nil
}

// ResetPassword validates the token and sets a new password for the user.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	tokenHash := tokenutil.Hash(token)

	resetToken, err := s.passwordResetRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return ErrInvalidToken
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return ErrTokenExpired
	}

	hashedPassword, err := auth.HashPassword(newPassword, s.config.Security.BcryptCost)
	if err != nil {
		s.logger.Error("failed to hash new password during reset", "user_id", resetToken.UserID, "error", err)
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, resetToken.UserID, hashedPassword); err != nil {
		s.logger.Error("failed to update password during reset", "user_id", resetToken.UserID, "error", err)
		return err
	}

	// Invalidate the token immediately after use.
	go func() {
		if err := s.passwordResetRepo.Delete(context.Background(), tokenHash); err != nil {
			s.logger.Error("failed to delete used password reset token", "token_hash", tokenHash, "error", err)
		}
	}()

	s.logger.Info("user password reset successfully", "user_id", resetToken.UserID)
	return nil
}

func (s *Service) Logout(ctx context.Context, w http.ResponseWriter, refreshTokenString string) error {
	s.clearAuthCookies(w)

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
	s.setAuthCookies(w, accessToken, refreshTokenString)

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

// setAuthCookies is a helper to set both access and refresh token cookies.
func (s *Service) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	// Access token cookie
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(s.config.JWT.AccessTokenExpiry),
		HttpOnly: true, 
		Secure:   s.config.Server.Env == "production",
		SameSite: http.SameSiteStrictMode,
	}

	// For non-production environments, use SameSiteLaxMode for broader compatibility
	if s.config.Server.Env != "production" {
		accessCookie.SameSite = http.SameSiteLaxMode
		// For cross-port communication on localhost, explicitly set the domain
		accessCookie.Domain = "localhost"
	}
	http.SetCookie(w, accessCookie)

	// Refresh token cookie
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/auth/refresh",
		Expires:  time.Now().Add(s.config.JWT.RefreshTokenExpiry),
		HttpOnly: true,
		Secure:   s.config.Server.Env == "production",
		SameSite: http.SameSiteStrictMode,
	}

	if s.config.Server.Env != "production" {
		refreshCookie.SameSite = http.SameSiteLaxMode
		refreshCookie.Domain = "localhost"
	}
	http.SetCookie(w, refreshCookie)
}

// clearAuthCookies clears both authentication cookies during logout.
func (s *Service) clearAuthCookies(w http.ResponseWriter) {
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, accessCookie)

	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, refreshCookie)
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
