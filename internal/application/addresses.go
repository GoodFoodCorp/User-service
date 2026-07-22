package application

import (
	"context"
	"errors"

	"goodfood/user-service/internal/domain"
)

// asDomainError is a small helper so use cases can branch on typed errors.
func asDomainError(err error, target **domain.Error) bool {
	return errors.As(err, target)
}

// ListMyAddresses returns the caller's saved delivery addresses.
func (uc *UseCases) ListMyAddresses(ctx context.Context, actor Actor) ([]domain.Address, error) {
	return uc.addresses.ListByUser(ctx, actor.UserID)
}

// AddAddress saves a new delivery address for the caller.
func (uc *UseCases) AddAddress(ctx context.Context, actor Actor, in domain.AddressInput) (*domain.Address, error) {
	address, err := domain.NewAddress(actor.UserID, in)
	if err != nil {
		return nil, err
	}
	// Only one default address per user.
	if address.IsDefault {
		if err := uc.addresses.ClearDefault(ctx, actor.UserID); err != nil {
			return nil, err
		}
	}
	if err := uc.addresses.Create(ctx, address); err != nil {
		return nil, err
	}
	return address, nil
}

// DeleteAddress removes one of the caller's own addresses.
func (uc *UseCases) DeleteAddress(ctx context.Context, actor Actor, id string) error {
	address, err := uc.addresses.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !address.IsOwnedBy(actor.UserID) {
		return domain.NewForbiddenError("this address belongs to another user")
	}
	return uc.addresses.Delete(ctx, id)
}
