package application

import (
	"context"

	"goodfood/user-service/internal/domain"
)

// ListMyFavorites returns every restaurant/dish the caller has favorited.
func (uc *UseCases) ListMyFavorites(ctx context.Context, actor Actor) ([]domain.Favorite, error) {
	return uc.favorites.ListByUser(ctx, actor.UserID)
}

// AddFavorite stars a restaurant or dish for the caller. Idempotent: starring
// twice is a no-op, not a conflict — the client can't easily know in advance.
func (uc *UseCases) AddFavorite(ctx context.Context, actor Actor, kind domain.FavoriteKind, targetID string) (*domain.Favorite, error) {
	favorite, err := domain.NewFavorite(actor.UserID, kind, targetID)
	if err != nil {
		return nil, err
	}
	if err := uc.favorites.Create(ctx, favorite); err != nil {
		return nil, err
	}
	return favorite, nil
}

// RemoveFavorite unstars a restaurant or dish. Idempotent: removing something
// that isn't favorited is a no-op.
func (uc *UseCases) RemoveFavorite(ctx context.Context, actor Actor, kind domain.FavoriteKind, targetID string) error {
	return uc.favorites.Delete(ctx, actor.UserID, kind, targetID)
}
