package auth

import "github.com/purushothdl/gochat-backend/pkg/errors"

var (
    ErrInvalidCredentials   = errors.New("INVALID_CREDENTIALS", "Invalid email or password", 401)
    ErrUserAlreadyExists    = errors.New("USER_ALREADY_EXISTS", "User with this email already exists", 409)
    ErrTokenExpired         = errors.New("TOKEN_EXPIRED", "Token has expired", 401)
    ErrInvalidToken         = errors.New("INVALID_TOKEN", "Invalid or malformed token", 401)
    ErrTooManyDevices       = errors.New("TOO_MANY_DEVICES", "Maximum number of devices reached", 429)
    ErrRefreshTokenNotFound = errors.New("REFRESH_TOKEN_NOT_FOUND", "Refresh token not found", 401)
)
