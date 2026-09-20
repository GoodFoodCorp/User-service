package domain

import "context"

type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*Profile, error)
	Create(ctx context.Context, profile *Profile) error
	Update(ctx context.Context, profile *Profile) error
	Exists(ctx context.Context, userID string) (bool, error)
}

type AddressRepository interface {
	ListByUser(ctx context.Context, userID string) ([]Address, error)
	GetByID(ctx context.Context, id string) (*Address, error)
	Create(ctx context.Context, address *Address) error
	Delete(ctx context.Context, id string) error
	ClearDefault(ctx context.Context, userID string) error
}

// FavoriteRepository and NotificationPreferencesRepository are declared next
// to their aggregate in favorite.go / notification_preferences.go.
