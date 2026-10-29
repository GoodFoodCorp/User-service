//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
)

// Run with: TEST_DATABASE_URL=postgres://... go test -tags integration ./internal/adapter/postgres/
func TestEraseUserDeletesEveryRowOfTheUserOnly(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	if err := Migrate(url); err != nil {
		t.Fatal(err)
	}
	pool, err := Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	alice, bob := uuid.NewString(), uuid.NewString()
	for _, u := range []string{alice, bob} {
		for _, q := range []string{
			`INSERT INTO profiles (user_id) VALUES ($1)`,
			`INSERT INTO addresses (id, user_id, street, city) VALUES (gen_random_uuid(), $1, '1 rue A', 'Paris')`,
		} {
			if _, err := pool.Exec(ctx, q, u); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := NewDataEraser(pool).EraseUser(ctx, alice); err != nil {
		t.Fatal(err)
	}

	count := func(table, user string) int {
		var n int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE user_id = $1`, user).Scan(&n)
		return n
	}
	for _, table := range erasableTables {
		if n := count(table, alice); n != 0 {
			t.Fatalf("%s still holds %d rows of the erased user", table, n)
		}
	}
	if count("profiles", bob) != 1 || count("addresses", bob) != 1 {
		t.Fatal("another user's data must be kept")
	}
}
