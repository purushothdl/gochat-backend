package response

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/purushothdl/gochat-backend/pkg/errors"
)

type APIResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
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

	var apiErr *APIError
	finalStatus := status

	// Handle both custom AppError and generic errors
	if appErr, ok := err.(*errors.AppError); ok {
		apiErr = &APIError{
            Code:    appErr.Code, 
            Message: appErr.Message,
        }
		finalStatus = appErr.Status
	} else {
		apiErr = &APIError{
            Code:    "INTERNAL_ERROR", 
            Message: err.Error()}
	}

	// Default to 500 if status is invalid
	if finalStatus == 0 {
		finalStatus = http.StatusInternalServerError
	}

	w.WriteHeader(finalStatus)
    
	response := APIResponse{
        Success: false, 
        Error:   apiErr,
    }

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// ErrorJSON sends a JSON response with a custom data payload for errors.
func ErrorJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := APIResponse{
		Success: false,
		Data:    data,
		Error: &APIError{
			Code:    "VALIDATION_ERROR",
			Message: "Request could not be processed due to invalid data.",
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

type MessageResponse struct {
	Message string `json:"message"`
}
