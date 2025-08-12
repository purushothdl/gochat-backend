package upload

import "context"

// UserProfileUpdater defines the contract the upload worker needs from the user domain
type UserProfileUpdater interface {
	UpdateUserImageURL(ctx context.Context, userID, imageURL string) error
}