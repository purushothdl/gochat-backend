package auth

import "time"

type LoginResponse struct {
    AccessToken string    `json:"access_token"`
    ExpiresAt   time.Time `json:"expires_at"`
    User        UserInfo  `json:"user"`
}

type RegisterResponse struct {
    AccessToken string    `json:"access_token"`
    ExpiresAt   time.Time `json:"expires_at"`
    User        UserInfo  `json:"user"`
}

type UserInfo struct {
    ID         string    `json:"id"`
    Email      string    `json:"email"`
    Name       string    `json:"name"`
    ImageURL   string    `json:"image_url"`
    IsVerified bool      `json:"is_verified"`
    CreatedAt  time.Time `json:"created_at"`
}

type RefreshTokenResponse struct {
    AccessToken string    `json:"access_token"`
    ExpiresAt   time.Time `json:"expires_at"`
}

type LogoutResponse struct {
    Message string `json:"message"`
}

type MeResponse struct {
    User UserInfo `json:"user"`
}
