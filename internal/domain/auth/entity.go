package auth

import "time"

type RefreshToken struct {
    ID         string    `json:"id"`
    UserID     string    `json:"user_id"`
    TokenHash  string    `json:"token_hash"`
    DeviceInfo string    `json:"device_info"`
    IPAddress  string    `json:"ip_address"`
    UserAgent  string    `json:"user_agent"`
    ExpiresAt  time.Time `json:"expires_at"`
    CreatedAt  time.Time `json:"created_at"`
    LastUsedAt time.Time `json:"last_used_at"`
}
