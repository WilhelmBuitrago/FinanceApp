package app

import (
	"context"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

// writeEventAndMovement creates a financial event and its movement (plus any
// components) through the given querier. Both transactional services (manual
// creation and review resolution) share this helper so the write exists once.
func writeEventAndMovement(ctx context.Context, q db.Querier, fe domain.FinancialEvent, mv domain.Movement, components []domain.MovementComponent) (domain.Movement, error) {
	repoDB := &repositories.DB{Querier: q}
	eventRepo := repositories.NewFinancialEventRepository(repoDB)
	movementRepo := repositories.NewMovementRepository(repoDB)

	if _, err := eventRepo.Create(ctx, fe); err != nil {
		return domain.Movement{}, err
	}
	created, err := movementRepo.Create(ctx, mv)
	if err != nil {
		return domain.Movement{}, err
	}
	for _, comp := range components {
		comp.MovementID = created.ID
		if _, err := movementRepo.CreateComponent(ctx, comp); err != nil {
			return domain.Movement{}, err
		}
	}
	return created, nil
}
