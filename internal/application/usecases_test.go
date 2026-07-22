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

func setup() *UseCases {
	return NewUseCases(newFakeProfiles(), newFakeAddresses())
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
