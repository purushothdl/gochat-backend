package user

import "github.com/purushothdl/gochat-backend/pkg/errors"

var (
    ErrUserNotFound        = errors.New("USER_NOT_FOUND", "User not found", 404)
    ErrEmailAlreadyExists  = errors.New("EMAIL_EXISTS", "Email already exists", 409)
    ErrInvalidPassword     = errors.New("INVALID_PASSWORD", "Current password is incorrect", 400)
    ErrNewPasswordSameAsOld = errors.New("NEW_PASSWORD_SAME_AS_OLD", "New password cannot be the same as the current password", 400)
    ErrUserNotActive       = errors.New("USER_NOT_ACTIVE", "User account is not active", 403)
    ErrPermissionDenied    = errors.New("PERMISSION_DENIED", "Permission denied", 403)

    ErrCannotBlockSelf = errors.New("CANNOT_BLOCK_SELF", "You cannot block yourself", 400)
	ErrAlreadyBlocked  = errors.New("ALREADY_BLOCKED", "This user is already blocked", 409)
	ErrNotBlocked      = errors.New("NOT_BLOCKED", "This user is not blocked", 404)
)
