package user

import "time"

type UserProfileResponse struct {
    ID          string               `json:"id"`
    Email       string               `json:"email"`
    Name        string               `json:"name"`
    ImageURL    string               `json:"image_url"`
    CreatedAt   time.Time            `json:"created_at"`
    IsVerified  bool                 `json:"is_verified"`
    LastLogin   *time.Time           `json:"last_login,omitempty"`
    Settings    UserSettingsResponse `json:"settings"`
}

type UserSettingsResponse struct {
    Theme                string `json:"theme"`
    NotificationsEnabled bool   `json:"notifications_enabled"`
    Language            string `json:"language"`
}

type UpdateProfileResponse struct {
    Message string               `json:"message"`
    User    UserProfileResponse  `json:"user"`
}

type UpdateSettingsResponse struct {
    Message  string               `json:"message"`
    Settings UserSettingsResponse `json:"settings"`
}
