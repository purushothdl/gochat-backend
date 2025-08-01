package user

import (
	"time"

	"github.com/google/uuid"
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
    Theme                string `json:"theme" default:"light"`
    NotificationsEnabled bool   `json:"notifications_enabled" default:"true"`
    Language             string `json:"language" default:"en"`
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


// Convert shared CreateUserData to domain User entity
func FromCreateUserData(data *types.CreateUserData) *User {
    return &User{
        ID:           uuid.New().String(), 
        Email:        data.Email,
        Name:         data.Name,
        PasswordHash: data.Password, 
        Settings:     UserSettings{
            Theme:                "light",
            NotificationsEnabled: true,
            Language:             "en",
        },
        IsVerified: false,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }
}

// Convert shared User to domain User entity  
func FromSharedUser(shared *types.User) *User {
    return &User{
        ID:         shared.ID,
        Email:      shared.Email,
        Name:       shared.Name,
        ImageURL:   shared.ImageURL,
        CreatedAt:  shared.CreatedAt,
        UpdatedAt:  shared.UpdatedAt,
        IsVerified: shared.IsVerified,
        LastLogin:  shared.LastLogin,
        // Password and Settings need separate handling
    }
}