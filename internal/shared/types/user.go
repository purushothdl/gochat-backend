package types

import "time"

// Shared user entity for cross-domain operations
type User struct {
    ID         string     `json:"id"`
    Email      string     `json:"email"`
    Name       string     `json:"name"`
    ImageURL   string     `json:"image_url"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
    IsVerified bool       `json:"is_verified"`
    LastLogin  *time.Time `json:"last_login,omitempty"`
}

type BasicUser struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    ImageURL string `json:"image_url"`
}

type CreateUserData struct {
    Email    string `json:"email"`
    Name     string `json:"name"`
    Password string `json:"password"`
}
