// internal/domain/message/handlers.go
package message

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	senderID, ok := authMiddleware.GetUserID(r.Context())
	h.logger.Info("sender ID from context", slog.String("user_id", senderID))
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}

	// Get the BasicUser from the context
	senderBasicUser, ok := authMiddleware.GetBasicUser(r.Context())
	
	// Log basic user details from context
	h.logger.Info("basic user from context",
		slog.String("user_id", senderBasicUser.ID),
		slog.String("name", senderBasicUser.Name),
		slog.String("image_url", senderBasicUser.ImageURL),
	)	
	
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.ErrInternalServer)
		return
	}



	roomID := chi.URLParam(r, "room_id")

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		response.JSON(w, http.StatusBadRequest, errs)
		return
	}

	msg, err := h.service.SendMessage(r.Context(), senderID, roomID, req.Content)
	if err != nil {
		response.Error(w, 0, err)
		return
	}

	msgWithFlag := &MessageWithSeenFlag{
		Message:      *msg,
		IsSeenByUser: true,
		User:         senderBasicUser, 
	}
	response.JSON(w, http.StatusCreated, msgWithFlag.ToResponse())
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}
	roomID := chi.URLParam(r, "room_id")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	before, _ := time.Parse(time.RFC3339Nano, r.URL.Query().Get("before_cursor"))

	messages, err := h.service.GetMessageHistory(r.Context(), userID, roomID, limit, before)
	if err != nil {
		response.Error(w, 0, err)
		return
	}

	resp := PaginatedMessagesResponse{
		Data:    make([]*MessageResponse, len(messages)),
		HasMore: len(messages) == limit,
	}
	for i, msg := range messages {
		resp.Data[i] = msg.ToResponse()
	}
	if resp.HasMore {
		oldestMsg := messages[len(messages)-1]
		resp.NextCursor = &oldestMsg.CreatedAt
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) EditMessage(w http.ResponseWriter, r *http.Request) {
	actorID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}
	messageID := chi.URLParam(r, "message_id")

	var req UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		response.JSON(w, http.StatusBadRequest, errs)
		return
	}

	if err := h.service.EditMessage(r.Context(), actorID, messageID, req.Content); err != nil {
		response.Error(w, 0, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "Message updated successfully"})
}

func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	actorID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}
	messageID := chi.URLParam(r, "message_id")
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "everyone"
	}

	if err := h.service.DeleteMessage(r.Context(), actorID, messageID, scope); err != nil {
		response.Error(w, 0, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MarkRoomRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}
	roomID := chi.URLParam(r, "room_id")

	var req ReadMarkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		response.JSON(w, http.StatusBadRequest, errs)
		return
	}

	if err := h.service.MarkRoomRead(r.Context(), userID, roomID, req.LastReadTimestamp); err != nil {
		response.Error(w, 0, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MarkMessagesSeen(w http.ResponseWriter, r *http.Request) {
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}

	var req BulkSeenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		response.JSON(w, http.StatusBadRequest, errs)
		return
	}

	if err := h.service.MarkMessagesAsSeen(r.Context(), userID, req.RoomID, req.MessageIDs); err != nil {
		response.Error(w, 0, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMessageReceipts(w http.ResponseWriter, r *http.Request) {
	userID, ok := authMiddleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return
	}
	messageID := chi.URLParam(r, "message_id")

	receipts, err := h.service.GetMessageReceipts(r.Context(), userID, messageID)
	if err != nil {
		response.Error(w, 0, err)
		return
	}

	response.JSON(w, http.StatusOK, receipts)
}
