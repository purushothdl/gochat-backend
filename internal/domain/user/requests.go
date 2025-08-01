package user

type UpdateProfileRequest struct {
    Name     *string `json:"name,omitempty" validate:"omitempty,min=2"`
    ImageURL *string `json:"image_url,omitempty" validate:"omitempty,url"`
}

type UpdateSettingsRequest struct {
    Theme                *string `json:"theme,omitempty" validate:"omitempty,oneof=light dark"`
    NotificationsEnabled *bool   `json:"notifications_enabled,omitempty"`
    Language            *string `json:"language,omitempty" validate:"omitempty,len=2"`
}

type ChangePasswordRequest struct {
    CurrentPassword string `json:"current_password" validate:"required"`
    NewPassword     string `json:"new_password" validate:"required,min=8"`
}
