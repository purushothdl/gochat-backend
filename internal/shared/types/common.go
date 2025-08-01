package types

import "time"

type ID string

type Timestamps struct {
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type PaginationParams struct {
    Page   int `json:"page"`
    Limit  int `json:"limit"`
    Offset int `json:"offset"`
}
