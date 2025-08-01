package auth

import (
    "encoding/json"
    "net/http"
    "strings"

    "github.com/purushothdl/gochat-backend/internal/shared/response"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, err)
        return
    }

    // Get device info from headers and request
    deviceInfo := r.Header.Get("User-Agent")
    ipAddress := h.getClientIP(r)
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

    // Get device info
    deviceInfo := r.Header.Get("User-Agent")
    ipAddress := h.getClientIP(r)
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
        response.Error(w, http.StatusUnauthorized, ErrRefreshTokenNotFound)
        return
    }

    deviceInfo := r.Header.Get("User-Agent")
    ipAddress := h.getClientIP(r)
    userAgent := r.Header.Get("User-Agent")

    result, err := h.service.RefreshToken(r.Context(), cookie.Value, deviceInfo, ipAddress, userAgent)
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

    if err := h.service.Logout(r.Context(), refreshToken); err != nil {
        response.Error(w, http.StatusInternalServerError, err)
        return
    }

    // Clear refresh token cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
        HttpOnly: true,
    })

    response.JSON(w, http.StatusOK, LogoutResponse{Message: "Logged out successfully"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
    // Get user ID from context (set by auth middleware)
    userID, ok := r.Context().Value("user_id").(string)
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

// Helper method to get client IP
func (h *Handler) getClientIP(r *http.Request) string {
    // Check X-Forwarded-For header first
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        // Get first IP if multiple
        ips := strings.Split(forwarded, ",")
        return strings.TrimSpace(ips[0])
    }

    // Check X-Real-IP header
    realIP := r.Header.Get("X-Real-IP")
    if realIP != "" {
        return realIP
    }

    // Fall back to RemoteAddr - but strip the port
    remoteAddr := r.RemoteAddr
    
    // Handle IPv6 addresses like [::1]:65384
    if strings.HasPrefix(remoteAddr, "[") {
        // IPv6 format: [::1]:port
        if idx := strings.LastIndex(remoteAddr, "]:"); idx != -1 {
            return remoteAddr[1:idx] 
        }
    }
    
    // Handle IPv4 addresses like 127.0.0.1:65384
    if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
        return remoteAddr[:idx] 
    }
    
    // If no port found, return as is
    return remoteAddr
}