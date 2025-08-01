package user

import (
    "encoding/json"
    "net/http"

    "github.com/purushothdl/gochat-backend/internal/shared/response"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value("user_id").(string)
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
    userID, ok := r.Context().Value("user_id").(string)
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
    userID, ok := r.Context().Value("user_id").(string)
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
    userID, ok := r.Context().Value("user_id").(string)
    if !ok {
        response.Error(w, http.StatusUnauthorized, ErrPermissionDenied)
        return
    }

    var req ChangePasswordRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, err)
        return
    }

    if err := h.service.ChangePassword(r.Context(), userID, req); err != nil {
        response.Error(w, http.StatusBadRequest, err)
        return
    }

    response.JSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}
