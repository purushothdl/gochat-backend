package user

import "github.com/purushothdl/gochat-backend/pkg/errors"

var (
    ErrUserNotFound        = errors.New("USER_NOT_FOUND", "User not found", 404)
    ErrEmailAlreadyExists  = errors.New("EMAIL_EXISTS", "Email already exists", 409)
    ErrInvalidPassword     = errors.New("INVALID_PASSWORD", "Current password is incorrect", 400)
    ErrUserNotActive       = errors.New("USER_NOT_ACTIVE", "User account is not active", 403)
    ErrPermissionDenied    = errors.New("PERMISSION_DENIED", "Permission denied", 403)
)
