package types

// RoomType is a shared definition for different room types.
type RoomType string

const (
	DirectRoom  RoomType = "DIRECT"
	PrivateRoom RoomType = "PRIVATE"
	PublicRoom  RoomType = "PUBLIC"
)

// MemberRole is a shared definition for member roles.
type MemberRole string

const (
	AdminRole   MemberRole = "ADMIN"
	RegularRole MemberRole = "MEMBER"
)

// RoomInfo contains the minimal room data needed by other domains.
type RoomInfo struct {
	ID              string
	Type            RoomType
	IsBroadcastOnly bool
}

// MembershipInfo contains the minimal membership data needed by other domains.
type MembershipInfo struct {
	RoomID string
	UserID string
	Role   MemberRole
}