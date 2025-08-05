// internal/domain/message/service.go
package message

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/shared/types"
	"github.com/purushothdl/gochat-backend/pkg/errors"
)

type Service struct {
	msgRepo  Repository
	roomProv RoomProvider
	userProv UserProvider
	config   *config.Config
	logger   *slog.Logger
	// redis publisher will be injected here
}

func NewService(msgRepo Repository, roomProv RoomProvider, userProv UserProvider, cfg *config.Config, logger *slog.Logger) *Service {
	return &Service{
		msgRepo:  msgRepo,
		roomProv: roomProv,
		userProv: userProv,
		config:   cfg,
		logger:   logger,
	}
}

func (s *Service) SendMessage(ctx context.Context, senderID, roomID, content string) (*Message, error) {
	membership, err := s.roomProv.GetMembershipInfo(ctx, roomID, senderID)
	if err != nil {
		return nil, err
	}

	targetRoom, err := s.roomProv.GetRoomInfo(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if targetRoom.IsBroadcastOnly && membership.Role != types.AdminRole {
		return nil, errors.New("BROADCAST_ONLY", "Only admins can send messages in this room.", 403)
	}

	msg := NewTextMessage(roomID, senderID, content)
	if err := s.msgRepo.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	s.logger.Info("message sent", "message_id", msg.ID, "room_id", roomID)
	// TODO: Publish real-time event via Redis
	return msg, nil
}

func (s *Service) GetMessageHistory(ctx context.Context, userID, roomID string, limit int, before time.Time) ([]*MessageWithSeenFlag, error) {
	if _, err := s.roomProv.GetMembershipInfo(ctx, roomID, userID); err != nil {
		return nil, err
	}
	if before.IsZero() {
		before = time.Now()
	}

	cursor := PaginationCursor{Timestamp: before, Limit: limit}
	return s.msgRepo.ListMessagesByRoom(ctx, roomID, userID, cursor)
}

func (s *Service) EditMessage(ctx context.Context, actorID, messageID, newContent string) error {
	msg, err := s.msgRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg.UserID == nil || *msg.UserID != actorID {
		return errors.New("NOT_OWNER", "You can only edit your own messages.", 403)
	}
	if time.Since(msg.CreatedAt) > (15 * time.Minute) {
		return ErrEditTimeExpired
	}

	// TODO: Publish real-time event via Redis
	return s.msgRepo.UpdateMessage(ctx, messageID, newContent)
}

func (s *Service) DeleteMessage(ctx context.Context, actorID, messageID, scope string) error {
	if scope == "me" {
		return s.msgRepo.DeleteMessageForUser(ctx, messageID, actorID)
	}

	msg, err := s.msgRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}
	membership, err := s.roomProv.GetMembershipInfo(ctx, msg.RoomID, actorID)
	if err != nil {
		return err
	}

	isSender := msg.UserID != nil && *msg.UserID == actorID
	isAdmin := membership.Role == types.AdminRole

	if !isSender && !isAdmin {
		return ErrDeleteNotAllowed
	}

	// TODO: Publish real-time event via Redis
	return s.msgRepo.SoftDeleteMessage(ctx, messageID)
}

func (s *Service) MarkRoomRead(ctx context.Context, userID, roomID string, timestamp time.Time) error {
	// TODO: Publish real-time event via Redis to update unread count on other devices
	return s.msgRepo.UpdateRoomReadMarker(ctx, roomID, userID, timestamp)
}

func (s *Service) MarkMessagesAsSeen(ctx context.Context, userID, roomID string, messageIDs []string) error {
	// TODO: Publish real-time event via Redis to notify sender of 'seen' status
	return s.msgRepo.CreateBulkReadReceipts(ctx, roomID, userID, messageIDs)
}

func (s *Service) GetMessageReceipts(ctx context.Context, actorID, messageID string) ([]*types.BasicUser, error) {
	msg, err := s.msgRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if _, err := s.roomProv.GetMembershipInfo(ctx, msg.RoomID, actorID); err != nil {
		return nil, err
	}
	return s.msgRepo.GetMessageReceipts(ctx, messageID)
}