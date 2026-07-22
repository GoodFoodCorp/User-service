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
	profiles  domain.ProfileRepository
	addresses domain.AddressRepository
}

func NewUseCases(profiles domain.ProfileRepository, addresses domain.AddressRepository) *UseCases {
	return &UseCases{profiles: profiles, addresses: addresses}
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
