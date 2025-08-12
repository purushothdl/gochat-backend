package contracts

import (
"context"
"io"
)

// FileStorage defines the interface for a cloud file storage system.
type FileStorage interface {
	Upload(ctx context.Context, key string, contentType string, body io.Reader, isPublic bool) error
	Download(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	GetPublicURL(key string) string
	FileExists(ctx context.Context, key string) (bool, error)
}