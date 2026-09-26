package application

import (
	"context"

	"goodfood/user-service/internal/domain"
)

// GetMyNotificationPreferences returns the caller's saved choice, or the
// product defaults if they never changed them.
func (uc *UseCases) GetMyNotificationPreferences(ctx context.Context, actor Actor) (*domain.NotificationPreferences, error) {
	prefs, err := uc.notifications.GetByUserID(ctx, actor.UserID)
	if err == nil {
		return prefs, nil
	}
	var derr *domain.Error
	if ok := asDomainError(err, &derr); ok && derr.Code == domain.ErrCodeNotFound {
		defaults := domain.DefaultNotificationPreferences(actor.UserID)
		return &defaults, nil
	}
	return nil, err
}

// UpdateMyNotificationPreferences saves the caller's channel choices.
func (uc *UseCases) UpdateMyNotificationPreferences(ctx context.Context, actor Actor, in domain.NotificationPreferencesInput) (*domain.NotificationPreferences, error) {
	prefs := &domain.NotificationPreferences{
		UserID:      actor.UserID,
		EmailOrders: in.EmailOrders,
		EmailPromos: in.EmailPromos,
		SmsOrders:   in.SmsOrders,
	}
	if err := uc.notifications.Upsert(ctx, prefs); err != nil {
		return nil, err
	}
	return prefs, nil
}
