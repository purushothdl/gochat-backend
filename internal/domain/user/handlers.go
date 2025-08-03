package user

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/gochat-backend/internal/shared/response"
	"github.com/purushothdl/gochat-backend/internal/shared/validator"
	authMiddleware "github.com/purushothdl/gochat-backend/internal/transport/http/middleware"
	"github.com/purushothdl/gochat-backend/pkg/errors"
)

type Handler struct {
	service   *Service
	logger    *slog.Logger
	validator *validator.Validator
}

func NewHandler(service *Service, logger *slog.Logger, validator *validator.Validator) *Handler {
	return &Handler{
		service:   service,
		logger:    logger,
		validator: validator,
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

    if errs := h.validator.Validate(req); errs != nil {
		response.ErrorJSON(w, http.StatusUnprocessableEntity, errs) 
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

    if errs := h.validator.Validate(req); errs != nil {
		response.ErrorJSON(w, http.StatusUnprocessableEntity, errs) 
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
    
	if errs := h.validator.Validate(req); errs != nil {
		response.ErrorJSON(w, http.StatusUnprocessableEntity, errs) 
		return
	}

	// The service layer handles all business logic and validation.
	if err := h.service.ChangePassword(r.Context(), userID, req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "Password changed successfully"})
}

// BlockUser handles POST /api/v1/users/me/blocks
func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	actorID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}

	var req BlockUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}
	if validationErrs := h.validator.Validate(req); validationErrs != nil {
		response.JSON(w, http.StatusBadRequest, validationErrs)
		return
	}

	if err := h.service.BlockUser(r.Context(), actorID, req.UserID); err != nil {
		response.Error(w, 0, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UnblockUser handles DELETE /api/v1/users/me/blocks/{user_id}
func (h *Handler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	actorID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}
	targetUserID := chi.URLParam(r, "user_id")

	if err := h.service.UnblockUser(r.Context(), actorID, targetUserID); err != nil {
		response.Error(w, 0, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListBlockedUsers handles GET /api/v1/users/me/blocks
func (h *Handler) ListBlockedUsers(w http.ResponseWriter, r *http.Request) {
	actorID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}

	blockedUsers, err := h.service.ListBlockedUsers(r.Context(), actorID)
	if err != nil {
		response.Error(w, 0, err)
		return
	}

	response.JSON(w, http.StatusOK, blockedUsers)
}