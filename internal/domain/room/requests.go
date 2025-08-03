package room

type CreateRoomRequest struct {
	Name string   `json:"name" validate:"required,min=3,max=50"`
	Type RoomType `json:"type" validate:"required,oneof=PRIVATE PUBLIC"`
}

type InviteUserRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type UpdateMemberRoleRequest struct {
	Role MemberRole `json:"role" validate:"required,oneof=ADMIN MEMBER"`
}

type UpdateRoomSettingsRequest struct {
	IsBroadcastOnly *bool `json:"is_broadcast_only"`
}