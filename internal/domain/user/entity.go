package user

import (
	"time"

	"github.com/purushothdl/gochat-backend/internal/shared/types"
)

type User struct {
    ID           string       `json:"id"`
    Email        string       `json:"email"`
    Name         string       `json:"name"`
    ImageURL     string       `json:"image_url"`
    PasswordHash string       `json:"-"`
    Settings     UserSettings `json:"settings"`
    CreatedAt    time.Time    `json:"created_at"`
    UpdatedAt    time.Time    `json:"updated_at"`
    IsVerified   bool         `json:"is_verified"`
    LastLogin    *time.Time   `json:"last_login,omitempty"`
    DeletedAt    *time.Time   `json:"-"`
}

type UserSettings struct {
    Theme                string `json:"theme"`
    NotificationsEnabled bool   `json:"notifications_enabled"`
    Language             string `json:"language"`
}

func NewDefaultUserSettings() UserSettings {
    return UserSettings{
        Theme:                "light",
        NotificationsEnabled: true,
        Language:             "en",
    }
}

func (u *User) ToSharedType() *types.User {
    return &types.User{
        ID:         u.ID,
        Email:      u.Email,
        Name:       u.Name,
        ImageURL:   u.ImageURL,
        CreatedAt:  u.CreatedAt,
        UpdatedAt:  u.UpdatedAt,
        IsVerified: u.IsVerified,
        LastLogin:  u.LastLogin,
    }
}

func (u *User) ToBasicUser() *types.BasicUser {
    return &types.BasicUser{
        ID:       u.ID,
        Name:     u.Name,
        ImageURL: u.ImageURL,
    }
}