// internal/domain/message/service.go
package message

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/contracts"
	"github.com/purushothdl/gochat-backend/internal/shared/types"
	"github.com/purushothdl/gochat-backend/pkg/errors"
)

type Service struct {
	msgRepo      Repository
	roomProv     RoomProvider
	userProv     UserProvider
	presenceProv PresenceProvider
	pubSub       contracts.PubSub
	config       *config.Config
	logger       *slog.Logger
}

func NewService(
	msgRepo Repository,
	roomProv RoomProvider,
	userProv UserProvider,
	presenceProv PresenceProvider,
	pubSub contracts.PubSub,
	cfg *config.Config,
	logger *slog.Logger,
) *Service {
	return &Service{
		msgRepo:      msgRepo,
		roomProv:     roomProv,
		userProv:     userProv,
		presenceProv: presenceProv,
		pubSub:       pubSub,
		config:       cfg,
		logger:       logger,
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

	// Publish the message to the Redis channel
	channel := fmt.Sprintf("room:%s:messages", roomID)
	messageJSON, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error("failed to marshal message for Redis", "error", err)
		return msg, nil
	}

	if err := s.pubSub.Publish(ctx, channel, string(messageJSON)); err != nil {
		s.logger.Error("failed to publish message to Redis", "error", err)
	}

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

func (s *Service) MarkMessagesAsSeen(ctx context.Context, userID, roomID string, messageIDs []string) error {
	// Step 1: Persist the individual receipts for the "blue tick" system.
	err := s.msgRepo.CreateBulkReadReceipts(ctx, roomID, userID, messageIDs)
	if err != nil {
		return err
	}
	// TODO: Publish real-time event to notify sender of 'seen' status.

	// Step 2: Determine if the user's unread count should be updated.
	latestTimestamp, err := s.msgRepo.GetLatestTimestampForMessages(ctx, messageIDs)
	if err != nil {
		s.logger.Error("failed to get latest timestamp for bulk seen", "error", err)
		return nil
	}

	// Step 3: Conditionally update the user's high-water mark.
	if err := s.msgRepo.UpdateRoomReadMarker(ctx, roomID, userID, *latestTimestamp); err != nil {
		s.logger.Error("failed to update room read marker during bulk seen", "error", err)
	}
	// TODO: Publish real-time event to update unread count on other devices.

	return nil
}

func (s *Service) GetMessageReceipts(ctx context.Context, actorID, messageID string) (*ReceiptDetailsResponse, error) {
	// 1. Authorize and get basic message/room info
	msg, err := s.msgRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	
	allMembers, err := s.roomProv.ListMembers(ctx, msg.RoomID)
	if err != nil {
		return nil, err
	}

	// 2. Get the "Read By" list (from the existing repository method)
	readByReceipts, err := s.msgRepo.GetMessageReceipts(ctx, messageID)
	if err != nil {
		return nil, err
	}

	// 3. Get the "Online" list (from our new provider)
	onlineUserIDs, err := s.presenceProv.GetOnlineUserIDs(ctx, msg.RoomID)
	if err != nil {
		// If Redis fails, we can gracefully degrade: return the "Read By" list and an empty "Delivered To" list.
		s.logger.Error("failed to get online users for receipts", "error", err)
		onlineUserIDs = []string{}
	}

	// 4. Perform the set logic designed
	readByMap := make(map[string]bool)
	for _, receipt := range readByReceipts {
		readByMap[receipt.User.ID] = true
	}

	onlineMap := make(map[string]bool)
	for _, userID := range onlineUserIDs {
		onlineMap[userID] = true
	}

	resp := &ReceiptDetailsResponse{
		ReadBy:      readByReceipts,
		DeliveredTo: []*types.ReceiptInfo{},
	}

	for _, member := range allMembers {
		// Skip the original sender
		if msg.UserID != nil && member.UserID == *msg.UserID {
			continue
		}

		// Check if the member has read the message
		if _, hasRead := readByMap[member.UserID]; hasRead {
			continue
		}

		// If not read, check if they are online to be marked as "Delivered"
		if _, isOnline := onlineMap[member.UserID]; isOnline {
			resp.DeliveredTo = append(resp.DeliveredTo, &types.ReceiptInfo{
				User: &types.BasicUser{
					ID:       member.UserID,
					Name:     member.Name,
					ImageURL: member.ImageURL,
				},
			})
		}
	}

	return resp, nil
}
