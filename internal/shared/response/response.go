package response

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/purushothdl/gochat-backend/pkg/errors"
)

type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

    response := APIResponse{
        Success: status < 400,
        Data:    data,
    }

    if err := json.NewEncoder(w).Encode(response); err != nil {
        log.Printf("Failed to encode JSON response: %v", err)
    }
}

func Error(w http.ResponseWriter, status int, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

    var apiErr *APIError
    
    // Check if it's our custom error type
    if appErr, ok := err.(*errors.AppError); ok {
        apiErr = &APIError{
            Code:    appErr.Code,
            Message: appErr.Message,
        }
        status = appErr.Status
        w.WriteHeader(status)
    } else {
        // Generic error
        apiErr = &APIError{
            Code:    "INTERNAL_ERROR",
            Message: err.Error(),
        }
    }

    response := APIResponse{
        Success: false,
        Error:   apiErr,
    }

    if err := json.NewEncoder(w).Encode(response); err != nil {
        log.Printf("Failed to encode error response: %v", err)
    }
}
