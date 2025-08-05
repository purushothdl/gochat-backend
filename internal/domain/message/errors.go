package message

import "github.com/purushothdl/gochat-backend/pkg/errors"

var (
	ErrMessageNotFound = errors.New("MESSAGE_NOT_FOUND", "The requested message was not found", 404)
	ErrEditTimeExpired = errors.New("EDIT_TIME_EXPIRED", "The time limit for editing this message has expired", 403)
	ErrDeleteNotAllowed = errors.New("DELETE_NOT_ALLOWED", "You do not have permission to delete this message", 403)
)