package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goodfood/user-service/internal/domain"
)

type fakeProfiles struct{ items map[string]*domain.Profile }

func newFakeProfiles() *fakeProfiles { return &fakeProfiles{items: map[string]*domain.Profile{}} }

func (f *fakeProfiles) GetByUserID(_ context.Context, userID string) (*domain.Profile, error) {
	p, ok := f.items[userID]
	if !ok {
		return nil, domain.NewNotFoundError("profile not found")
	}
	cp := *p
	return &cp, nil
}

func (f *fakeProfiles) Create(_ context.Context, p *domain.Profile) error {
	cp := *p
	f.items[p.UserID] = &cp
	return nil
}

func (f *fakeProfiles) Update(_ context.Context, p *domain.Profile) error {
	f.items[p.UserID] = p
	return nil
}

func (f *fakeProfiles) Exists(_ context.Context, userID string) (bool, error) {
	_, ok := f.items[userID]
	return ok, nil
}

type fakeAddresses struct{ items map[string]*domain.Address }

func newFakeAddresses() *fakeAddresses { return &fakeAddresses{items: map[string]*domain.Address{}} }

func (f *fakeAddresses) ListByUser(_ context.Context, userID string) ([]domain.Address, error) {
	out := []domain.Address{}
	for _, a := range f.items {
		if a.UserID == userID {
			out = append(out, *a)
		}
	}
	return out, nil
}

func (f *fakeAddresses) GetByID(_ context.Context, id string) (*domain.Address, error) {
	a, ok := f.items[id]
	if !ok {
		return nil, domain.NewNotFoundError("address not found")
	}
	cp := *a
	return &cp, nil
}

func (f *fakeAddresses) Create(_ context.Context, a *domain.Address) error {
	cp := *a
	f.items[a.ID] = &cp
	return nil
}

func (f *fakeAddresses) Delete(_ context.Context, id string) error {
	delete(f.items, id)
	return nil
}

func (f *fakeAddresses) ClearDefault(_ context.Context, userID string) error {
	for _, a := range f.items {
		if a.UserID == userID {
			a.IsDefault = false
		}
	}
	return nil
}

type fakeFavorites struct{ items map[string]*domain.Favorite }

func newFakeFavorites() *fakeFavorites { return &fakeFavorites{items: map[string]*domain.Favorite{}} }

func favoriteKey(userID string, kind domain.FavoriteKind, targetID string) string {
	return userID + "|" + string(kind) + "|" + targetID
}

func (f *fakeFavorites) ListByUser(_ context.Context, userID string) ([]domain.Favorite, error) {
	out := []domain.Favorite{}
	for _, fav := range f.items {
		if fav.UserID == userID {
			out = append(out, *fav)
		}
	}
	return out, nil
}

func (f *fakeFavorites) Exists(_ context.Context, userID string, kind domain.FavoriteKind, targetID string) (bool, error) {
	_, ok := f.items[favoriteKey(userID, kind, targetID)]
	return ok, nil
}

func (f *fakeFavorites) Create(_ context.Context, fav *domain.Favorite) error {
	key := favoriteKey(fav.UserID, fav.Kind, fav.TargetID)
	if existing, ok := f.items[key]; ok {
		*fav = *existing
		return nil
	}
	cp := *fav
	f.items[key] = &cp
	return nil
}

func (f *fakeFavorites) Delete(_ context.Context, userID string, kind domain.FavoriteKind, targetID string) error {
	delete(f.items, favoriteKey(userID, kind, targetID))
	return nil
}

type fakeNotificationPreferences struct {
	items map[string]*domain.NotificationPreferences
}

func newFakeNotificationPreferences() *fakeNotificationPreferences {
	return &fakeNotificationPreferences{items: map[string]*domain.NotificationPreferences{}}
}

func (f *fakeNotificationPreferences) GetByUserID(_ context.Context, userID string) (*domain.NotificationPreferences, error) {
	p, ok := f.items[userID]
	if !ok {
		return nil, domain.NewNotFoundError("notification preferences not found")
	}
	cp := *p
	return &cp, nil
}

func (f *fakeNotificationPreferences) Upsert(_ context.Context, p *domain.NotificationPreferences) error {
	cp := *p
	f.items[p.UserID] = &cp
	return nil
}

func setup() *UseCases {
	return NewUseCases(newFakeProfiles(), newFakeAddresses(), newFakeFavorites(), newFakeNotificationPreferences())
}

var alice = Actor{UserID: "user-1", RoleSlugs: []string{"user"}}
var bob = Actor{UserID: "user-2", RoleSlugs: []string{"user"}}

func TestEnsureProfileIsIdempotent(t *testing.T) {
	uc := setup()
	require.NoError(t, uc.EnsureProfile(context.Background(), alice.UserID))
	require.NoError(t, uc.EnsureProfile(context.Background(), alice.UserID))

	profile, err := uc.GetMyProfile(context.Background(), alice)
	require.NoError(t, err)
	assert.Equal(t, alice.UserID, profile.UserID)
}

func TestGetMyProfileCreatesOnTheFly(t *testing.T) {
	uc := setup()
	// No profile yet (user registered before this service existed).
	profile, err := uc.GetMyProfile(context.Background(), alice)
	require.NoError(t, err)
	assert.Equal(t, alice.UserID, profile.UserID)
}

func TestUpdateMyProfile(t *testing.T) {
	uc := setup()
	profile, err := uc.UpdateMyProfile(context.Background(), alice,
		domain.ProfileInput{FirstName: " Marie ", LastName: "Dupont", Phone: "0601020304"})
	require.NoError(t, err)
	assert.Equal(t, "Marie", profile.FirstName, "input should be trimmed")
	assert.Equal(t, "Dupont", profile.LastName)
}

func TestAddressLifecycleAndOwnership(t *testing.T) {
	uc := setup()

	addr, err := uc.AddAddress(context.Background(), alice,
		domain.AddressInput{Street: "12 rue de Paris", City: "Paris", ZipCode: "75001", IsDefault: true})
	require.NoError(t, err)
	assert.Equal(t, "Domicile", addr.Label, "a default label is applied")
	assert.Equal(t, "12 rue de Paris, 75001, Paris", addr.FullAddress())

	list, err := uc.ListMyAddresses(context.Background(), alice)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Bob sees nothing of Alice's, and cannot delete her address.
	bobList, _ := uc.ListMyAddresses(context.Background(), bob)
	assert.Empty(t, bobList)

	err = uc.DeleteAddress(context.Background(), bob, addr.ID)
	var derr *domain.Error
	require.ErrorAs(t, err, &derr)
	assert.Equal(t, domain.ErrCodeForbidden, derr.Code)

	require.NoError(t, uc.DeleteAddress(context.Background(), alice, addr.ID))
}

func TestAddressValidation(t *testing.T) {
	uc := setup()
	_, err := uc.AddAddress(context.Background(), alice, domain.AddressInput{City: "Paris"})
	var derr *domain.Error
	require.ErrorAs(t, err, &derr)
	assert.Equal(t, domain.ErrCodeValidation, derr.Code)
}

func TestOnlyOneDefaultAddress(t *testing.T) {
	uc := setup()
	first, err := uc.AddAddress(context.Background(), alice,
		domain.AddressInput{Street: "1 rue A", City: "Paris", IsDefault: true})
	require.NoError(t, err)
	_, err = uc.AddAddress(context.Background(), alice,
		domain.AddressInput{Street: "2 rue B", City: "Lyon", IsDefault: true})
	require.NoError(t, err)

	list, _ := uc.ListMyAddresses(context.Background(), alice)
	defaults := 0
	for _, a := range list {
		if a.IsDefault {
			defaults++
		}
	}
	assert.Equal(t, 1, defaults, "adding a new default clears the previous one")
	assert.NotEmpty(t, first.ID)
}

func TestAddFavoriteIsIdempotent(t *testing.T) {
	uc := setup()
	first, err := uc.AddFavorite(context.Background(), alice, domain.FavoriteKindDish, "dish-1")
	require.NoError(t, err)
	second, err := uc.AddFavorite(context.Background(), alice, domain.FavoriteKindDish, "dish-1")
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID, "starring twice keeps the original row")

	list, _ := uc.ListMyFavorites(context.Background(), alice)
	assert.Len(t, list, 1)

	require.NoError(t, uc.RemoveFavorite(context.Background(), alice, domain.FavoriteKindDish, "dish-1"))
	list, _ = uc.ListMyFavorites(context.Background(), alice)
	assert.Empty(t, list)
}

func TestNotificationPreferencesDefaultThenUpdate(t *testing.T) {
	uc := setup()
	prefs, err := uc.GetMyNotificationPreferences(context.Background(), alice)
	require.NoError(t, err)
	assert.True(t, prefs.EmailOrders, "orders email is opt-out, not opt-in")

	updated, err := uc.UpdateMyNotificationPreferences(context.Background(), alice,
		domain.NotificationPreferencesInput{EmailOrders: false, EmailPromos: false, SmsOrders: true})
	require.NoError(t, err)
	assert.False(t, updated.EmailOrders)
	assert.True(t, updated.SmsOrders)
}
