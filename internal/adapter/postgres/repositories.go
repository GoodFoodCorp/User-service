package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"goodfood/user-service/internal/domain"
)

type ProfileRepository struct {
	pool *pgxpool.Pool
}

func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

func (r *ProfileRepository) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	var p domain.Profile
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, first_name, last_name, phone, created_at, updated_at
		 FROM profiles WHERE user_id = $1`, userID).
		Scan(&p.UserID, &p.FirstName, &p.LastName, &p.Phone, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NewNotFoundError("profile not found")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProfileRepository) Create(ctx context.Context, p *domain.Profile) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO profiles (user_id, first_name, last_name, phone, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (user_id) DO NOTHING`,
		p.UserID, p.FirstName, p.LastName, p.Phone, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *ProfileRepository) Update(ctx context.Context, p *domain.Profile) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE profiles SET first_name=$1, last_name=$2, phone=$3, updated_at=$4 WHERE user_id=$5`,
		p.FirstName, p.LastName, p.Phone, p.UpdatedAt, p.UserID)
	return err
}

func (r *ProfileRepository) Exists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM profiles WHERE user_id = $1)`, userID).Scan(&exists)
	return exists, err
}

type AddressRepository struct {
	pool *pgxpool.Pool
}

func NewAddressRepository(pool *pgxpool.Pool) *AddressRepository {
	return &AddressRepository{pool: pool}
}

func (r *AddressRepository) ListByUser(ctx context.Context, userID string) ([]domain.Address, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, label, street, zip_code, city, is_default, created_at
		 FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []domain.Address{}
	for rows.Next() {
		var a domain.Address
		if err := rows.Scan(&a.ID, &a.UserID, &a.Label, &a.Street, &a.ZipCode,
			&a.City, &a.IsDefault, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *AddressRepository) GetByID(ctx context.Context, id string) (*domain.Address, error) {
	var a domain.Address
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, label, street, zip_code, city, is_default, created_at
		 FROM addresses WHERE id = $1`, id).
		Scan(&a.ID, &a.UserID, &a.Label, &a.Street, &a.ZipCode, &a.City, &a.IsDefault, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NewNotFoundError("address not found")
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AddressRepository) Create(ctx context.Context, a *domain.Address) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO addresses (id, user_id, label, street, zip_code, city, is_default, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.UserID, a.Label, a.Street, a.ZipCode, a.City, a.IsDefault, a.CreatedAt)
	return err
}

func (r *AddressRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM addresses WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NewNotFoundError("address not found")
	}
	return nil
}

func (r *AddressRepository) ClearDefault(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE addresses SET is_default = FALSE WHERE user_id = $1`, userID)
	return err
}
