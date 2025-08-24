package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/purushothdl/gochat-backend/internal/shared/response"
	"github.com/purushothdl/gochat-backend/internal/shared/validator"
	authMiddleware "github.com/purushothdl/gochat-backend/internal/transport/http/middleware"
	"github.com/purushothdl/gochat-backend/pkg/utils/httputil"
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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	// Validate the request body
	if errs := h.validator.Validate(req); errs != nil {
		response.ErrorJSON(w, http.StatusUnprocessableEntity, errs)
		return
	}

	// Get device info from headers and request
	deviceInfo := r.Header.Get("User-Agent")
	ipAddress := httputil.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	result, err := h.service.Register(r.Context(), req, w, deviceInfo, ipAddress, userAgent)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ErrorJSON(w, http.StatusUnprocessableEntity, errs)
		return
	}

	// Get device info
	deviceInfo := r.Header.Get("User-Agent")
	ipAddress := httputil.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	result, err := h.service.Login(r.Context(), req, w, deviceInfo, ipAddress, userAgent)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")

	if err != nil {
		h.logger.Info("Failed to get refresh token from cookie", "error", err)
		response.Error(w, http.StatusUnauthorized, ErrRefreshTokenNotFound)
		return
	}

	deviceInfo := r.Header.Get("User-Agent")
	ipAddress := httputil.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	result, err := h.service.RefreshAccessToken(r.Context(), cookie.Value, deviceInfo, ipAddress, userAgent)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	refreshToken := ""
	if err == nil {
		refreshToken = cookie.Value
	}

	if err := h.service.Logout(r.Context(), w, refreshToken); err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "Logged out successfully"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, ErrInvalidToken)
		return
	}

	result, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ErrorJSON(w, http.StatusUnprocessableEntity, errs)
		return
	}

	// The service handles all logic. We deliberately don't return an error
	// to prevent user enumeration attacks.
	if err := h.service.ForgotPassword(r.Context(), req.Email); err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{
		Message: "If an account with that email exists, we have sent a password reset link.",
	})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ErrorJSON(w, http.StatusUnprocessableEntity, errs)
		return
	}

	if err := h.service.ResetPassword(r.Context(), req.Token, req.Password); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "Your password has been reset successfully."})
}

func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, ErrInvalidToken)
		return
	}

	devices, err := h.service.ListDevices(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, ListDevicesResponse{Devices: devices})
}

func (h *Handler) LogoutDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, ErrInvalidToken)
		return
	}

	var req LogoutDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	// Get the current device ID from the context
	currentDeviceID, ok := authMiddleware.GetDeviceID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, ErrInvalidToken)
		return
	}

	// If the user is revoking the current session, raise an error
	if req.DeviceID == currentDeviceID {
		response.ErrorJSON(w, http.StatusForbidden, "Please log out to delete the current session.")
		return
	}

	if err := h.service.LogoutDevice(r.Context(), userID, req.DeviceID); err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "Device logged out successfully"})
}

func (h *Handler) LogoutAllDevices(w http.ResponseWriter, r *http.Request) {
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, ErrInvalidToken)
		return
	}

	// Clear auth cookies
	h.service.clearAuthCookies(w)

	if err := h.service.LogoutAllDevices(r.Context(), userID); err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "Logged out from all devices successfully"})
}
