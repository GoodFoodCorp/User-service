package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// FavoriteKind is what a favorite points to. Both targets live in other
// services (franchise-service, menu-service) — this service only stores the
// pointer, keyed by the customer.
type FavoriteKind string

const (
	FavoriteKindRestaurant FavoriteKind = "restaurant"
	FavoriteKindDish       FavoriteKind = "dish"
)

func (k FavoriteKind) valid() bool {
	return k == FavoriteKindRestaurant || k == FavoriteKindDish
}

type Favorite struct {
	ID        string
	UserID    string
	Kind      FavoriteKind
	TargetID  string
	CreatedAt time.Time
}

func NewFavorite(userID string, kind FavoriteKind, targetID string) (*Favorite, error) {
	if userID == "" {
		return nil, NewValidationError("user id is required")
	}
	if !kind.valid() {
		return nil, NewValidationError("favorite kind must be 'restaurant' or 'dish'")
	}
	if targetID == "" {
		return nil, NewValidationError("target id is required")
	}
	return &Favorite{
		ID:        uuid.NewString(),
		UserID:    userID,
		Kind:      kind,
		TargetID:  targetID,
		CreatedAt: time.Now().UTC(),
	}, nil
}

type FavoriteRepository interface {
	ListByUser(ctx context.Context, userID string) ([]Favorite, error)
	Exists(ctx context.Context, userID string, kind FavoriteKind, targetID string) (bool, error)
	Create(ctx context.Context, favorite *Favorite) error
	Delete(ctx context.Context, userID string, kind FavoriteKind, targetID string) error
}
