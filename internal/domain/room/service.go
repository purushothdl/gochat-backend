// internal/domain/room/service.go
package room

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/pkg/errors"
)

type Service struct {
	roomRepo Repository
	userProv UserProvider
	config   *config.Config
	logger   *slog.Logger
}

func NewService(repo Repository, userProv UserProvider, cfg *config.Config, logger *slog.Logger) *Service {
	return &Service{
		roomRepo: repo,
		userProv: userProv,
		config:   cfg,
		logger:   logger,
	}
}

// CreateRoom handles the creation of a new room and assigns the creator as an admin.
func (s *Service) CreateRoom(ctx context.Context, creatorID string, req CreateRoomRequest) (*Room, error) {
	newRoom := NewPrivateRoom(req.Name)
	newRoom.Type = req.Type

	err := s.roomRepo.CreateRoom(ctx, newRoom)
	if err != nil {
		return nil, fmt.Errorf("service failed to create room: %w", err)
	}

	// The creator automatically becomes an admin of the new room.
	adminMembership := &RoomMembership{
		RoomID: newRoom.ID,
		UserID: creatorID,
		Role:   AdminRole,
	}

	err = s.roomRepo.CreateMembership(ctx, adminMembership)
	if err != nil {
		return nil, fmt.Errorf("service failed to create admin membership: %w", err)
	}

	s.logger.Info("new room created", "room_id", newRoom.ID, "user_id", creatorID)
	return newRoom, nil
}

// InviteUser handles inviting a new user to a private room.
func (s *Service) InviteUser(ctx context.Context, inviterID, roomID, inviteeID string) error {
	// 1. Verify the person sending the invite is an admin.
	inviterMembership, err := s.roomRepo.FindMembership(ctx, roomID, inviterID)
	if err != nil {
		return err // Could be ErrNotMember or a db error
	}
	if inviterMembership.Role != AdminRole {
		return ErrNotAdmin
	}

	// 2. Check if the user to be invited actually exists.
	exists, err := s.userProv.ExistsByID(ctx, inviteeID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}
    
    // 3. Check if the user is already a member.
    _, err = s.roomRepo.FindMembership(ctx, roomID, inviteeID)
    if err == nil {
        return ErrAlreadyInRoom
    }
    if err != ErrNotMember {
        // A different database error occurred
        return err
    }

	// 4. Create the new membership.
	newMembership := &RoomMembership{
		RoomID: roomID,
		UserID: inviteeID,
		Role:   RegularRole,
	}
	if err := s.roomRepo.CreateMembership(ctx, newMembership); err != nil {
		return fmt.Errorf("service failed to create membership for invitee: %w", err)
	}

	s.logger.Info("user invited to room", "room_id", roomID, "inviter_id", inviterID, "invitee_id", inviteeID)
	return nil
}

// JoinPublicRoom allows a user to become a member of a public room.
func (s *Service) JoinPublicRoom(ctx context.Context, userID, roomID string) error {
	targetRoom, err := s.roomRepo.FindRoomByID(ctx, roomID)
	if err != nil {
		return err 
	}

	if targetRoom.Type != PublicRoom {
		return errors.New("NOT_PUBLIC", "This room is not public.", 403)
	}
    
    // Check if user is already a member to prevent constraint violation errors.
	if _, err := s.roomRepo.FindMembership(ctx, roomID, userID); err != ErrNotMember {
		if err == nil {
			return ErrAlreadyInRoom
		}
		return err 
	}
    
    // Create the new membership with a default 'MEMBER' role.
	newMembership := &RoomMembership{
		RoomID: roomID,
		UserID: userID,
		Role:   RegularRole,
	}
	return s.roomRepo.CreateMembership(ctx, newMembership)
}

// ListUserRooms retrieves a list of rooms the given user is a member of.
func (s *Service) ListUserRooms(ctx context.Context, userID string) ([]*Room, error) {
	return s.roomRepo.ListUserRooms(ctx, userID)
}

// ListPublicRooms retrieves all rooms that are open for any user to join.
func (s *Service) ListPublicRooms(ctx context.Context) ([]*Room, error) {
	return s.roomRepo.ListPublicRooms(ctx)
}

// ListMembers retrieves the member list for a room, ensuring the requester is a member.
func (s *Service) ListMembers(ctx context.Context, requesterID, roomID string) ([]*MemberDetail, error) {
	// Authorization: Verify the user requesting the list is a member of the room.
	if _, err := s.roomRepo.FindMembership(ctx, roomID, requesterID); err != nil {
		return nil, err 
	}
	return s.roomRepo.ListMembers(ctx, roomID)
}

// UpdateMemberRole changes the role of a user within a room.
func (s *Service) UpdateMemberRole(ctx context.Context, actorID, roomID, targetUserID string, newRole MemberRole) error {
	// 1. Authorize: Ensure the actor is an admin.
	actorMembership, err := s.roomRepo.FindMembership(ctx, roomID, actorID)
	if err != nil {
		return err
	}
	if actorMembership.Role != AdminRole {
		return ErrNotAdmin
	}
    
    // Prevent an admin from demoting themselves if they are the last one.
    if actorID == targetUserID && newRole != AdminRole {
        adminCount, err := s.roomRepo.CountAdmins(ctx, roomID)
        if err != nil {
            return err
        }
        if adminCount <= 1 {
            return errors.New("LAST_ADMIN", "You cannot demote yourself as the last admin.", 403)
        }
    }

	// 2. Find the target user's membership to update.
	targetMembership, err := s.roomRepo.FindMembership(ctx, roomID, targetUserID)
	if err != nil {
		return err 
	}

	// 3. Update and save.
	targetMembership.Role = newRole
	return s.roomRepo.UpdateMembership(ctx, targetMembership)
}

// RemoveMember kicks a user from a room.
func (s *Service) RemoveMember(ctx context.Context, actorID, roomID, targetUserID string) error {
	// 1. Authorize: Ensure the actor is an admin.
	actorMembership, err := s.roomRepo.FindMembership(ctx, roomID, actorID)
	if err != nil {
		return err
	}
	if actorMembership.Role != AdminRole {
		return ErrNotAdmin
	}
    
    // An admin cannot kick themselves. They must use the LeaveRoom functionality.
    if actorID == targetUserID {
        return errors.New("CANNOT_KICK_SELF", "Admins cannot kick themselves. Please use the 'Leave Room' option.", 400)
    }

	// 2. Delete the target user's membership.
	return s.roomRepo.DeleteMembership(ctx, roomID, targetUserID)
}

// LeaveRoom allows a user to remove themselves from a room.
func (s *Service) LeaveRoom(ctx context.Context, userID, roomID string) error {
	// 1. Check if user is the last admin.
	membership, err := s.roomRepo.FindMembership(ctx, roomID, userID)
	if err != nil {
		return err
	}

	if membership.Role == AdminRole {
		adminCount, err := s.roomRepo.CountAdmins(ctx, roomID)
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return errors.New("LAST_ADMIN_LEAVE", "You cannot leave as the last admin. Promote another member first.", 403)
		}
	}

	// 2. Delete the user's membership.
	return s.roomRepo.DeleteMembership(ctx, roomID, userID)
}

func (s *Service) UpdateRoomSettings(ctx context.Context, actorID, roomID string, req UpdateRoomSettingsRequest) (*Room, error) {
	// 1. Authorize: Ensure the actor is an admin.
	actorMembership, err := s.roomRepo.FindMembership(ctx, roomID, actorID)
	if err != nil {
		return nil, err
	}
	if actorMembership.Role != AdminRole {
		return nil, ErrNotAdmin
	}

	// 2. Fetch the current state of the room.
	targetRoom, err := s.roomRepo.FindRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// 3. Apply changes if they were provided in the request.
	if req.IsBroadcastOnly != nil {
		targetRoom.IsBroadcastOnly = *req.IsBroadcastOnly
	}

	// 4. Persist the changes.
	if err := s.roomRepo.UpdateRoom(ctx, targetRoom); err != nil {
		return nil, err
	}

	return targetRoom, nil
}