package room

import "github.com/purushothdl/gochat-backend/pkg/errors"

var (
	ErrNotAdmin       = errors.New("NOT_ADMIN", "You must be an admin to perform this action", 403)
	ErrAlreadyInRoom  = errors.New("ALREADY_IN_ROOM", "User is already a member of this room", 409)
	ErrRoomNotFound   = errors.New("ROOM_NOT_FOUND", "The requested room was not found", 404)
	ErrUserNotFound   = errors.New("USER_TO_INVITE_NOT_FOUND", "The user you are trying to invite does not exist", 404)
	ErrNotMember      = errors.New("NOT_A_MEMBER", "You are not a member of this room", 403)
)