package errors

import (
    "fmt"
    "net/http"
)

type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}

// Common errors
var (
    ErrNotFound          = &AppError{"NOT_FOUND", "Resource not found", http.StatusNotFound}
    ErrUnauthorized      = &AppError{"UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized}
    ErrForbidden         = &AppError{"FORBIDDEN", "Forbidden access", http.StatusForbidden}
    ErrBadRequest        = &AppError{"BAD_REQUEST", "Bad request", http.StatusBadRequest}
    ErrInternalServer    = &AppError{"INTERNAL_SERVER", "Internal server error", http.StatusInternalServerError}
    ErrExternalServer    = &AppError{"EXTERNAL_SERVER", "External server error", http.StatusBadGateway}
    ErrValidationFailed  = &AppError{"VALIDATION_FAILED", "Validation failed", http.StatusBadRequest}
)

func New(code, message string, status int) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Status:  status,
    }
}

func Wrap(err error, message string) error {
    return fmt.Errorf("%s: %w", message, err)
}
