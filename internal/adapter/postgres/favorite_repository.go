package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"goodfood/user-service/internal/domain"
)

type FavoriteRepository struct {
	pool *pgxpool.Pool
}

func NewFavoriteRepository(pool *pgxpool.Pool) *FavoriteRepository {
	return &FavoriteRepository{pool: pool}
}

func (r *FavoriteRepository) ListByUser(ctx context.Context, userID string) ([]domain.Favorite, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, kind, target_id, created_at
		 FROM favorites WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []domain.Favorite{}
	for rows.Next() {
		var f domain.Favorite
		if err := rows.Scan(&f.ID, &f.UserID, &f.Kind, &f.TargetID, &f.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

func (r *FavoriteRepository) Exists(ctx context.Context, userID string, kind domain.FavoriteKind, targetID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND kind = $2 AND target_id = $3)`,
		userID, kind, targetID).Scan(&exists)
	return exists, err
}

// Create is idempotent: starring the same target twice keeps the original row
// (and its original created_at) instead of erroring.
func (r *FavoriteRepository) Create(ctx context.Context, f *domain.Favorite) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO favorites (id, user_id, kind, target_id, created_at)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (user_id, kind, target_id)
		 DO UPDATE SET kind = favorites.kind
		 RETURNING id, created_at`,
		f.ID, f.UserID, f.Kind, f.TargetID, f.CreatedAt).Scan(&f.ID, &f.CreatedAt)
}

func (r *FavoriteRepository) Delete(ctx context.Context, userID string, kind domain.FavoriteKind, targetID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM favorites WHERE user_id = $1 AND kind = $2 AND target_id = $3`, userID, kind, targetID)
	return err
}
