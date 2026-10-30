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
	AvatarURL string
	Age       *int
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
	Age       *int
}

func (p *Profile) Update(in ProfileInput) error {
	if len(in.Phone) > 30 {
		return NewValidationError("phone number is too long")
	}
	if in.Age != nil && (*in.Age < 0 || *in.Age > 120) {
		return NewValidationError("age must be between 0 and 120")
	}
	p.FirstName = strings.TrimSpace(in.FirstName)
	p.LastName = strings.TrimSpace(in.LastName)
	p.Phone = strings.TrimSpace(in.Phone)
	p.Age = in.Age
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// SetAvatar records the public URL of the uploaded avatar image.
func (p *Profile) SetAvatar(url string) {
	p.AvatarURL = url
	p.UpdatedAt = time.Now().UTC()
}

// Address is a saved delivery address belonging to one user.
type Address struct {
	ID      string
	UserID  string
	Label   string
	Street  string
	ZipCode string
	City    string
	// CountryCode is the ISO 3166-1 alpha-2 code (an EU member state).
	CountryCode string
	IsDefault   bool
	CreatedAt   time.Time
}

type AddressInput struct {
	Label   string
	Street  string
	ZipCode string
	City    string
	// CountryCode defaults to FR when empty.
	CountryCode string
	IsDefault   bool
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
	country := strings.ToUpper(strings.TrimSpace(in.CountryCode))
	if country == "" {
		country = DefaultCountryCode
	}
	if _, served := euCountryNames[country]; !served {
		return nil, NewValidationError("deliveries are limited to the European Union")
	}
	label := strings.TrimSpace(in.Label)
	if label == "" {
		label = "Domicile"
	}
	return &Address{
		ID:          uuid.NewString(),
		UserID:      userID,
		Label:       label,
		Street:      strings.TrimSpace(in.Street),
		ZipCode:     strings.TrimSpace(in.ZipCode),
		City:        strings.TrimSpace(in.City),
		CountryCode: country,
		IsDefault:   in.IsDefault,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// IsOwnedBy reports whether the address belongs to the given user.
func (a *Address) IsOwnedBy(userID string) bool { return a.UserID == userID }

// FullAddress renders the address on one line, country included so it is
// geocoded in the right member state (used by the checkout).
func (a *Address) FullAddress() string {
	out := []string{}
	if a.Street != "" {
		out = append(out, a.Street)
	}
	if locality := strings.TrimSpace(a.ZipCode + " " + a.City); locality != "" {
		out = append(out, locality)
	}
	if name := euCountryNames[a.CountryCode]; name != "" {
		out = append(out, name)
	}
	return strings.Join(out, ", ")
}

// DefaultCountryCode applies to addresses saved without a country.
const DefaultCountryCode = "FR"

// euCountryNames lists the countries deliveries are made in (the EU member
// states), with their French names.
var euCountryNames = map[string]string{
	"AT": "Autriche", "BE": "Belgique", "BG": "Bulgarie", "HR": "Croatie", "CY": "Chypre",
	"CZ": "Tchéquie", "DK": "Danemark", "EE": "Estonie", "FI": "Finlande", "FR": "France",
	"DE": "Allemagne", "GR": "Grèce", "HU": "Hongrie", "IE": "Irlande", "IT": "Italie",
	"LV": "Lettonie", "LT": "Lituanie", "LU": "Luxembourg", "MT": "Malte", "NL": "Pays-Bas",
	"PL": "Pologne", "PT": "Portugal", "RO": "Roumanie", "SK": "Slovaquie", "SI": "Slovénie",
	"ES": "Espagne", "SE": "Suède",
}
