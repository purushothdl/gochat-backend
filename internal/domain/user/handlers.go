package user

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/purushothdl/gochat-backend/internal/shared/response"
    authMiddleware "github.com/purushothdl/gochat-backend/internal/transport/http/middleware" 
)

type Handler struct {
	service *Service
	logger  *slog.Logger 
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger, 
	}
}


func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
    userID, ok := authMiddleware.GetUserID(r.Context())
    if !ok {
        response.Error(w, http.StatusUnauthorized, ErrPermissionDenied)
        return
    }

    profile, err := h.service.GetProfile(r.Context(), userID)
    if err != nil {
        response.Error(w, http.StatusNotFound, err)
        return
    }

    response.JSON(w, http.StatusOK, profile)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
    userID, ok := authMiddleware.GetUserID(r.Context())
    if !ok {
        response.Error(w, http.StatusUnauthorized, ErrPermissionDenied)
        return
    }

    var req UpdateProfileRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, err)
        return
    }

    result, err := h.service.UpdateProfile(r.Context(), userID, req)
    if err != nil {
        response.Error(w, http.StatusBadRequest, err)
        return
    }

    response.JSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
    userID, ok := authMiddleware.GetUserID(r.Context())
    if !ok {
        response.Error(w, http.StatusUnauthorized, ErrPermissionDenied)
        return
    }

    var req UpdateSettingsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, err)
        return
    }

    result, err := h.service.UpdateSettings(r.Context(), userID, req)
    if err != nil {
        response.Error(w, http.StatusBadRequest, err)
        return
    }

    response.JSON(w, http.StatusOK, result)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		h.logger.Error("user_id not found in context for authenticated route")
		response.Error(w, http.StatusUnauthorized, ErrPermissionDenied)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode change password request", "error", err, "user_id", userID)
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	// The service layer handles all business logic and validation.
	if err := h.service.ChangePassword(r.Context(), userID, req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "Password changed successfully"})
}