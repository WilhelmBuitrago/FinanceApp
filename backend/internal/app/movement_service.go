package app

import (
	"context"
	"database/sql"
	"fmt"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

// MovementService orchestrates movement writes that must be atomic: financial
// event + movement + components commit together or not at all. The import
// flow already worked transactionally; this extends the same pattern to the
// manual creation path (and is the foundation for credit card breakdowns,
// where a bad component must never leave a half-written movement behind).
type MovementService struct {
	db *sql.DB
}

func NewMovementService(db *sql.DB) *MovementService {
	return &MovementService{db: db}
}

// CreateWithEvent inserts a financial event, its movement and the movement's
// components in a single transaction.
func (s *MovementService) CreateWithEvent(ctx context.Context, fe domain.FinancialEvent, mv domain.Movement, components []domain.MovementComponent) (created domain.Movement, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Movement{}, fmt.Errorf("begin tx: %w", err)
	}
	// Named return: any failure below arms the rollback.
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	repoDB := &repositories.DB{Querier: db.New(tx)}
	eventRepo := repositories.NewFinancialEventRepository(repoDB)
	movementRepo := repositories.NewMovementRepository(repoDB)

	if _, err := eventRepo.Create(ctx, fe); err != nil {
		return domain.Movement{}, fmt.Errorf("create financial event: %w", err)
	}
	created, err = movementRepo.Create(ctx, mv)
	if err != nil {
		return domain.Movement{}, fmt.Errorf("create movement: %w", err)
	}
	for _, comp := range components {
		comp.MovementID = created.ID
		if _, err := movementRepo.CreateComponent(ctx, comp); err != nil {
			return domain.Movement{}, fmt.Errorf("create movement component: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return domain.Movement{}, fmt.Errorf("commit: %w", err)
	}
	return created, nil
}
