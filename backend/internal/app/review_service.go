package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

// ReviewService turns a human-verified review entry into real financial data:
// the user checked the raw against the PDF and supplied the correct values.
type ReviewService struct {
	db *sql.DB
}

func NewReviewService(db *sql.DB) *ReviewService {
	return &ReviewService{db: db}
}

// ListEntries returns the pending review queue.
func (s *ReviewService) ListEntries(ctx context.Context) ([]repositories.ReviewEntry, error) {
	repo := repositories.NewReviewRepository(&repositories.DB{Querier: db.New(s.db)})
	return repo.ListEntries(ctx)
}

// Resolve writes the approved movement and marks the source entry resolved in
// one transaction. The entry is marked FIRST: a second submission finds it no
// longer pending and fails before any money is written.
func (s *ReviewService) Resolve(ctx context.Context, entryID string, description string, amountMinor int64, date string, eventType string, accountID string, categoryID *string) (created domain.Movement, err error) {
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

	now := time.Now().UTC().Format(time.RFC3339)
	repoDB := &repositories.DB{Querier: db.New(tx)}
	reviewRepo := repositories.NewReviewRepository(repoDB)

	if err = reviewRepo.Resolve(ctx, entryID, &description, &amountMinor, &date, now); err != nil {
		return domain.Movement{}, err
	}

	fe := domain.FinancialEvent{
		ID:                 uuid.NewString(),
		EventType:          eventType,
		EventDate:          &date,
		Status:             "ACTIVE",
		VerificationStatus: "VERIFIED", // a human checked this one
		Description:        &description,
		CategoryID:         categoryID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	mv := domain.Movement{
		ID:               uuid.NewString(),
		FinancialEventID: fe.ID,
		AccountID:        accountID,
		MovementDate:     date,
		AmountMinor:      amountMinor,
		Description:      &description,
		Status:           "ACTIVE",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	created, err = writeEventAndMovement(ctx, db.New(tx), fe, mv, nil)
	if err != nil {
		return domain.Movement{}, fmt.Errorf("write approved movement: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return domain.Movement{}, fmt.Errorf("commit: %w", err)
	}
	return created, nil
}
