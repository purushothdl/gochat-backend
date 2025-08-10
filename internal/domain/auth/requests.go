package auth

import "time"

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=6"`
	Password string `json:"password" validate:"required,min=8,password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,password"`
}

type PasswordResetToken struct {
	UserID    string
	ExpiresAt time.Time
}

type LogoutDeviceRequest struct {
	DeviceID string `json:"device_id" validate:"required"`
}
