package user

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/shared/types"
	"github.com/purushothdl/gochat-backend/pkg/auth"
	pointer "github.com/purushothdl/gochat-backend/pkg/utils/pointer"
)

type Service struct {
	repo   Repository
	config *config.Config
	logger *slog.Logger 
}

func NewService(repo Repository, cfg *config.Config, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		config: cfg,
		logger: logger, 
	}
}

func (s *Service) GetProfile(ctx context.Context, userID string) (*UserProfileResponse, error) {
    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        return nil, ErrUserNotFound
    }
    
    return &UserProfileResponse{
        ID:         user.ID,
        Email:      user.Email,
        Name:       user.Name,
        ImageURL:   user.ImageURL,
        CreatedAt:  user.CreatedAt,
        IsVerified: user.IsVerified,
        LastLogin:  user.LastLogin,
        Settings: UserSettingsResponse{
            Theme:                user.Settings.Theme,
            NotificationsEnabled: user.Settings.NotificationsEnabled,
            Language:            user.Settings.Language,
        },
    }, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*UpdateProfileResponse, error) {
    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        return nil, ErrUserNotFound
    }
    
    // Update fields if provided
	pointer.UpdatePointerField(&user.Name, req.Name)
	pointer.UpdatePointerField(&user.ImageURL, req.ImageURL)
    
    if err := s.repo.Update(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to update user: %w", err)
    }
    
    profile, err := s.GetProfile(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    return &UpdateProfileResponse{
        Message: "Profile updated successfully",
        User:    *profile,
    }, nil
}

func (s *Service) UpdateSettings(ctx context.Context, userID string, req UpdateSettingsRequest) (*UpdateSettingsResponse, error) {
    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        return nil, ErrUserNotFound
    }
    
    // Update settings if provided
    pointer.UpdatePointerField(&user.Settings.Theme, req.Theme)
    pointer.UpdatePointerField(&user.Settings.NotificationsEnabled, req.NotificationsEnabled)
    pointer.UpdatePointerField(&user.Settings.Language, req.Language) 

    if err := s.repo.Update(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to update user settings: %w", err)
    }
    
    return &UpdateSettingsResponse{
        Message: "Settings updated successfully",
        Settings: UserSettingsResponse{
            Theme:                user.Settings.Theme,
            NotificationsEnabled: user.Settings.NotificationsEnabled,
            Language:            user.Settings.Language,
        },
    }, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
	if req.CurrentPassword == req.NewPassword {
		return ErrNewPasswordSameAsOld
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Warn("user not found for password change", "user_id", userID)
		return ErrUserNotFound
	}

	// Verify the user's current password.
	if !auth.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		return ErrInvalidPassword
	}

	// Hash the new password.
	newHashedPassword, err := auth.HashPassword(req.NewPassword, s.config.Security.BcryptCost)
	if err != nil {
		s.logger.Error("failed to hash new password", "user_id", userID, "error", err)
		return fmt.Errorf("could not process new password: %w", err)
	}

	user.PasswordHash = newHashedPassword

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update password in database", "user_id", userID, "error", err)
		return fmt.Errorf("failed to save new password: %w", err)
	}

	s.logger.Info("user password changed successfully", "user_id", userID)
	return nil
}

func (s *Service) DeleteAccount(ctx context.Context, userID string) error {
    if err := s.repo.Delete(ctx, userID); err != nil {
        return fmt.Errorf("failed to delete account: %w", err)
    }
    
    return nil
}

// BlockUser creates a block relationship from the actor to the target user.
func (s *Service) BlockUser(ctx context.Context, actorID, targetUserID string) error {
	if actorID == targetUserID {
		return ErrCannotBlockSelf
	}

	// Ensure the user being blocked actually exists.
	exists, err := s.repo.ExistsByID(ctx, targetUserID) 
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	return s.repo.BlockUser(ctx, actorID, targetUserID)
}

// UnblockUser removes a block relationship.
func (s *Service) UnblockUser(ctx context.Context, actorID, targetUserID string) error {
	return s.repo.UnblockUser(ctx, actorID, targetUserID)
}

// ListBlockedUsers returns a list of basic user profiles the actor has blocked.
func (s *Service) ListBlockedUsers(ctx context.Context, actorID string) ([]*types.BasicUser, error) {
	return s.repo.ListBlockedUsers(ctx, actorID)
}