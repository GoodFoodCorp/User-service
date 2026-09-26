package domain

import (
	"context"
	"time"
)

// NotificationPreferences controls which channels a customer wants to hear
// from. Defaults (all-on for orders, promos opt-in) apply until the customer
// saves their own choice.
type NotificationPreferences struct {
	UserID      string
	EmailOrders bool
	EmailPromos bool
	SmsOrders   bool
	UpdatedAt   time.Time
}

func DefaultNotificationPreferences(userID string) NotificationPreferences {
	return NotificationPreferences{
		UserID:      userID,
		EmailOrders: true,
		EmailPromos: true,
		SmsOrders:   false,
		UpdatedAt:   time.Now().UTC(),
	}
}

type NotificationPreferencesInput struct {
	EmailOrders bool
	EmailPromos bool
	SmsOrders   bool
}

type NotificationPreferencesRepository interface {
	GetByUserID(ctx context.Context, userID string) (*NotificationPreferences, error)
	Upsert(ctx context.Context, prefs *NotificationPreferences) error
}
