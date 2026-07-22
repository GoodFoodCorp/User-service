package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Profile holds the personal data of a customer. Identity/credentials stay in
// auth-service: this service only owns the profile, keyed by the auth user id.
type Profile struct {
	UserID    string
	FirstName string
	LastName  string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewProfile(userID string) (*Profile, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, NewValidationError("user id is required")
	}
	now := time.Now().UTC()
	return &Profile{UserID: userID, CreatedAt: now, UpdatedAt: now}, nil
}

type ProfileInput struct {
	FirstName string
	LastName  string
	Phone     string
}

func (p *Profile) Update(in ProfileInput) error {
	if len(in.Phone) > 30 {
		return NewValidationError("phone number is too long")
	}
	p.FirstName = strings.TrimSpace(in.FirstName)
	p.LastName = strings.TrimSpace(in.LastName)
	p.Phone = strings.TrimSpace(in.Phone)
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// Address is a saved delivery address belonging to one user.
type Address struct {
	ID        string
	UserID    string
	Label     string
	Street    string
	ZipCode   string
	City      string
	IsDefault bool
	CreatedAt time.Time
}

type AddressInput struct {
	Label     string
	Street    string
	ZipCode   string
	City      string
	IsDefault bool
}

func NewAddress(userID string, in AddressInput) (*Address, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, NewValidationError("user id is required")
	}
	if strings.TrimSpace(in.Street) == "" {
		return nil, NewValidationError("street is required")
	}
	if strings.TrimSpace(in.City) == "" {
		return nil, NewValidationError("city is required")
	}
	label := strings.TrimSpace(in.Label)
	if label == "" {
		label = "Domicile"
	}
	return &Address{
		ID:        uuid.NewString(),
		UserID:    userID,
		Label:     label,
		Street:    strings.TrimSpace(in.Street),
		ZipCode:   strings.TrimSpace(in.ZipCode),
		City:      strings.TrimSpace(in.City),
		IsDefault: in.IsDefault,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// IsOwnedBy reports whether the address belongs to the given user.
func (a *Address) IsOwnedBy(userID string) bool { return a.UserID == userID }

// FullAddress renders the address on one line (used by the checkout).
func (a *Address) FullAddress() string {
	parts := []string{a.Street, a.ZipCode, a.City}
	out := []string{}
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, ", ")
}
