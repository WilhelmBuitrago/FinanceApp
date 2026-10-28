package app

import (
	"context"
	"database/sql"
	"fmt"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
)

// MovementService orchestrates movement writes that must be atomic: financial
// event + movement + components commit together or not at all. The import
// flow already worked transactionally; this extends the same pattern to the
// manual creation path (and is the foundation for credit card breakdowns,
// where a bad component must never leave a half-written movement behind).
type MovementService struct {
	db         *sql.DB
	duplicates DuplicateSubmitter
}

func NewMovementService(db *sql.DB, duplicates DuplicateSubmitter) *MovementService {
	return &MovementService{db: db, duplicates: duplicates}
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
			_ = tx.Rollback()
		}
	}()

	created, err = writeEventAndMovement(ctx, db.New(tx), fe, mv, components)
	if err != nil {
		return domain.Movement{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Movement{}, fmt.Errorf("commit: %w", err)
	}
	return created, nil
}
