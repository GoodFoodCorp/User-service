package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"goodfood/user-service/internal/domain"
)

type NotificationPreferencesRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationPreferencesRepository(pool *pgxpool.Pool) *NotificationPreferencesRepository {
	return &NotificationPreferencesRepository{pool: pool}
}

func (r *NotificationPreferencesRepository) GetByUserID(ctx context.Context, userID string) (*domain.NotificationPreferences, error) {
	var p domain.NotificationPreferences
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, email_orders, email_promos, sms_orders, updated_at
		 FROM notification_preferences WHERE user_id = $1`, userID).
		Scan(&p.UserID, &p.EmailOrders, &p.EmailPromos, &p.SmsOrders, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NewNotFoundError("notification preferences not found")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *NotificationPreferencesRepository) Upsert(ctx context.Context, p *domain.NotificationPreferences) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO notification_preferences (user_id, email_orders, email_promos, sms_orders, updated_at)
		 VALUES ($1,$2,$3,$4,now())
		 ON CONFLICT (user_id) DO UPDATE SET
		   email_orders = EXCLUDED.email_orders,
		   email_promos = EXCLUDED.email_promos,
		   sms_orders   = EXCLUDED.sms_orders,
		   updated_at   = EXCLUDED.updated_at
		 RETURNING updated_at`,
		p.UserID, p.EmailOrders, p.EmailPromos, p.SmsOrders).Scan(&p.UpdatedAt)
}
