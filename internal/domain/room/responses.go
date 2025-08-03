package room

import "time"


type RoomResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      RoomType  `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type MemberResponse struct {
	UserID   string     `json:"user_id"`
	Role     MemberRole `json:"role"`
	Name     string     `json:"name"`
	ImageURL string     `json:"image_url"`
}

func (r *Room) ToResponse() *RoomResponse {
	return &RoomResponse{
		ID:        r.ID,
		Name:      r.Name,
		Type:      r.Type,
		CreatedAt: r.CreatedAt,
	}
}

func (d *MemberDetail) ToResponse() *MemberResponse {
	return &MemberResponse{
		UserID:   d.UserID,
		Role:     d.Role,
		Name:     d.Name,
		ImageURL: d.ImageURL,
	}
}