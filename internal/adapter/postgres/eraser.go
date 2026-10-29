package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DataEraser deletes every row this service holds about a user.
type DataEraser struct {
	pool *pgxpool.Pool
}

func NewDataEraser(pool *pgxpool.Pool) *DataEraser { return &DataEraser{pool: pool} }

// erasableTables are all the tables keyed by user_id.
var erasableTables = []string{"addresses", "favorites", "notification_preferences", "user_settings", "profiles"}

func (e *DataEraser) EraseUser(ctx context.Context, userID string) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, table := range erasableTables {
		// Table names come from the constant list above, never from input.
		if _, err := tx.Exec(ctx, `DELETE FROM `+table+` WHERE user_id = $1`, userID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
