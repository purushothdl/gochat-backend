package room

import (
	"time"

	"github.com/purushothdl/gochat-backend/internal/shared/types"
)


type RoomResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Type            RoomType  `json:"type"`
	IsBroadcastOnly bool      `json:"is_broadcast_only"`
	CreatedAt       time.Time `json:"created_at"`
}

type MemberResponse struct {
	UserID   string            `json:"user_id"`
	Role     types.MemberRole  `json:"role"`
	Name     string            `json:"name"`
	ImageURL string            `json:"image_url"`
}

func (r *Room) ToResponse() *RoomResponse {
	return &RoomResponse{
		ID:              r.ID,
		Name:            r.Name,
		Type:            r.Type,
		IsBroadcastOnly: r.IsBroadcastOnly,
		CreatedAt:       r.CreatedAt,
	}
}

func MemberDetailToResponse(d *types.MemberDetail) *MemberResponse {
	return &MemberResponse{
		UserID:   d.UserID,
		Role:     types.MemberRole(d.Role), 
		Name:     d.Name,
		ImageURL: d.ImageURL,
	}
}