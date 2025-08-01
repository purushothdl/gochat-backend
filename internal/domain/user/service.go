package user

import (
    "context"
    "fmt"
    
    "github.com/purushothdl/gochat-backend/pkg/auth"
    "github.com/purushothdl/gochat-backend/internal/config"
)

type Service struct {
    repo         Repository
    config       *config.Config
    // notification NotificationService 
}

func NewService(repo Repository, cfg *config.Config) *Service {
    return &Service{
        repo:   repo,
        config: cfg,
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
    if req.Name != nil {
        user.Name = *req.Name
    }
    if req.ImageURL != nil {
        user.ImageURL = *req.ImageURL
    }
    
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
    if req.Theme != nil {
        user.Settings.Theme = *req.Theme
    }
    if req.NotificationsEnabled != nil {
        user.Settings.NotificationsEnabled = *req.NotificationsEnabled
    }
    if req.Language != nil {
        user.Settings.Language = *req.Language
    }
    
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
    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        return ErrUserNotFound
    }
    
    // Check current password
    if !auth.CheckPassword(req.CurrentPassword, user.PasswordHash) {
        return ErrInvalidPassword
    }
    
    // Hash new password
    newHashedPassword, err := auth.HashPassword(req.NewPassword, s.config.Security.BcryptCost)
    if err != nil {
        return fmt.Errorf("failed to hash new password: %w", err)
    }
    
    user.PasswordHash = newHashedPassword
    
    if err := s.repo.Update(ctx, user); err != nil {
        return fmt.Errorf("failed to update password: %w", err)
    }
    
    return nil
}

func (s *Service) DeleteAccount(ctx context.Context, userID string) error {
    if err := s.repo.Delete(ctx, userID); err != nil {
        return fmt.Errorf("failed to delete account: %w", err)
    }
    
    return nil
}
