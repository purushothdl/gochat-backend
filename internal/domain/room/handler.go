// internal/domain/room/handlers.go
package room

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

func NewHandler(service *Service, logger *slog.Logger, v *validator.Validator) *Handler {
	return &Handler{
		service:   service,
		logger:    logger,
		validator: v,
	}
}

// CreateRoom handles POST /api/v1/rooms
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	creatorID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}
	if validationErrs := h.validator.Validate(req); validationErrs != nil {
		response.ErrorJSON(w, http.StatusUnprocessableEntity, validationErrs) 
		return
	}

	newRoom, err := h.service.CreateRoom(r.Context(), creatorID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusCreated, newRoom.ToResponse())
}

// InviteUser hadnles POST /api/v1/rooms/{room_id}/invite
func (h *Handler) InviteUser(w http.ResponseWriter, r *http.Request) {
	inviterID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}
	roomID := chi.URLParam(r, "room_id")

	var req InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}
	if validationErrs := h.validator.Validate(req); validationErrs != nil {
		response.ErrorJSON(w, http.StatusUnprocessableEntity, validationErrs) 
		return
	}

	err := h.service.InviteUser(r.Context(), inviterID, roomID, req.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err) 
		return
	}

	response.JSON(w, http.StatusCreated, response.MessageResponse{Message: "User invited successfully"})
}

// ListUserRooms handles GET /api/v1/rooms
func (h *Handler) ListUserRooms(w http.ResponseWriter, r *http.Request) {
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}

	rooms, err := h.service.ListUserRooms(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	roomResponses := make([]*RoomResponse, len(rooms))
	for i, room := range rooms {
		roomResponses[i] = room.ToResponse()
	}

	response.JSON(w, http.StatusOK, roomResponses)
}

// ListPublicRooms handles GET /api/v1/rooms/public
func (h *Handler) ListPublicRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.service.ListPublicRooms(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	roomResponses := make([]*RoomResponse, len(rooms))
	for i, room := range rooms {
		roomResponses[i] = room.ToResponse()
	}

	response.JSON(w, http.StatusOK, roomResponses)
}

// ListMembers handles GET /api/v1/rooms/{room_id}/members
func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}
	roomID := chi.URLParam(r, "room_id")

	members, err := h.service.ListMembers(r.Context(), requesterID, roomID)
	if err != nil {
		response.Error(w, 0, err) 
		return
	}

	memberResponses := make([]*MemberResponse, len(members))
	for i, member := range members {
		memberResponses[i] = member.ToResponse()
	}

	response.JSON(w, http.StatusOK, memberResponses)
}

// JoinPublicRoom handles POST /api/v1/rooms/{room_id}/join
func (h *Handler) JoinPublicRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}
	roomID := chi.URLParam(r, "room_id")

	if err := h.service.JoinPublicRoom(r.Context(), userID, roomID); err != nil {
		response.Error(w, 0, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "Successfully joined room"})
}