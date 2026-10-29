package domain

import (
	"context"
	"io"
)

// AvatarStore keeps avatar images outside the service's pods (object storage
// in production), so they survive restarts and are shared by every replica.
type AvatarStore interface {
	// Put stores the image under key and returns its public URL.
	Put(ctx context.Context, key, contentType string, body io.Reader, size int64) (string, error)
	// Delete removes the image previously returned for that URL; unknown URLs
	// are ignored.
	Delete(ctx context.Context, url string) error
}

// UserDataEraser deletes everything this service stores about a user (right
// to erasure), in one transaction.
type UserDataEraser interface {
	EraseUser(ctx context.Context, userID string) error
}
