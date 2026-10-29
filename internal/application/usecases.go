package application

import (
	"context"

	"goodfood/user-service/internal/domain"
)

// Actor is the authenticated caller (from the auth-service JWT).
type Actor struct {
	UserID    string
	RoleSlugs []string
}

type UseCases struct {
	profiles      domain.ProfileRepository
	addresses     domain.AddressRepository
	favorites     domain.FavoriteRepository
	notifications domain.NotificationPreferencesRepository
	avatars       domain.AvatarStore
	eraser        domain.UserDataEraser
	settings      interface {
		GetSettings(context.Context, string) (*domain.UserSettings, error)
		UpdateSettings(context.Context, string, domain.SettingsInput) (*domain.UserSettings, error)
	}
}

// Deps lists the ports the use cases need.
type Deps struct {
	Profiles      domain.ProfileRepository
	Addresses     domain.AddressRepository
	Favorites     domain.FavoriteRepository
	Notifications domain.NotificationPreferencesRepository
	Avatars       domain.AvatarStore
	Eraser        domain.UserDataEraser
}

func NewUseCases(d Deps) *UseCases {
	settings, _ := d.Profiles.(interface {
		GetSettings(context.Context, string) (*domain.UserSettings, error)
		UpdateSettings(context.Context, string, domain.SettingsInput) (*domain.UserSettings, error)
	})
	return &UseCases{profiles: d.Profiles, addresses: d.Addresses, favorites: d.Favorites, notifications: d.Notifications,
		avatars: d.Avatars, eraser: d.Eraser, settings: settings}
}

func (uc *UseCases) GetMySettings(ctx context.Context, actor Actor) (*domain.UserSettings, error) {
	if uc.settings == nil { return nil, domain.NewValidationError("settings unavailable") }
	return uc.settings.GetSettings(ctx, actor.UserID)
}

func (uc *UseCases) UpdateMySettings(ctx context.Context, actor Actor, in domain.SettingsInput) (*domain.UserSettings, error) {
	if uc.settings == nil { return nil, domain.NewValidationError("settings unavailable") }
	return uc.settings.UpdateSettings(ctx, actor.UserID, in)
}

// EnsureProfile creates a blank profile if none exists. Called by auth-service
// right after registration (POST /internal/profiles) — idempotent.
func (uc *UseCases) EnsureProfile(ctx context.Context, userID string) error {
	exists, err := uc.profiles.Exists(ctx, userID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	profile, err := domain.NewProfile(userID)
	if err != nil {
		return err
	}
	return uc.profiles.Create(ctx, profile)
}

// GetMyProfile returns the caller's profile, creating it on the fly for users
// registered before this service existed.
func (uc *UseCases) GetMyProfile(ctx context.Context, actor Actor) (*domain.Profile, error) {
	profile, err := uc.profiles.GetByUserID(ctx, actor.UserID)
	if err == nil {
		return profile, nil
	}
	var derr *domain.Error
	if ok := asDomainError(err, &derr); ok && derr.Code == domain.ErrCodeNotFound {
		if createErr := uc.EnsureProfile(ctx, actor.UserID); createErr != nil {
			return nil, createErr
		}
		return uc.profiles.GetByUserID(ctx, actor.UserID)
	}
	return nil, err
}

// UpdateMyProfile edits the caller's own profile.
func (uc *UseCases) UpdateMyProfile(ctx context.Context, actor Actor, in domain.ProfileInput) (*domain.Profile, error) {
	profile, err := uc.GetMyProfile(ctx, actor)
	if err != nil {
		return nil, err
	}
	if err := profile.Update(in); err != nil {
		return nil, err
	}
	if err := uc.profiles.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

