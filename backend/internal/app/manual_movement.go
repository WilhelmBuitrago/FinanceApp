package app

import (
	"context"
	"fmt"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/log"
	"financeapp/backend/internal/repositories"
)

// DuplicateSubmitter parks a colliding manual movement in the unified
// review queue. ReviewService satisfies it; it is injected because the
// api package cannot import app (app imports api), so the transaction
// owner cannot name the reviewer either.
type DuplicateSubmitter interface {
	Submit(ctx context.Context, kind string, payload any, sourceEntryID *string) (string, error)
}

// CreateManual owns the manual-creation policy: a collision on the
// natural key (same account, date, amount, and a description empty on
// either side) parks the movement in the review queue and reports a
// conflict — the human decides, never a silent drop; otherwise the
// financial event and the movement are written atomically.
//
// A failure to park never fails the request: the collision is still
// reported as a conflict, and the parked item is simply lost this once
// (the human re-enters it). A review-queue outage must not turn a
// duplicate into a fabricated success.
func (s *MovementService) CreateManual(ctx context.Context, cmd domain.ManualMovementCommand) (domain.ManualMovementOutcome, error) {
	if !cmd.Force {
		repoDB := &repositories.DB{Querier: db.New(s.db)}
		movementRepo := repositories.NewMovementRepository(repoDB)
		collisions, err := movementRepo.FindByNaturalKey(ctx, cmd.AccountID, cmd.Date, cmd.AmountMinor, cmd.Description)
		if err != nil {
			return domain.ManualMovementOutcome{}, err
		}
		if len(collisions) > 0 {
			m := collisions[0]
			log.Info(fmt.Sprintf("[MOVEMENT CREATE] duplicate of %s", m.ID))
			itemID, parkErr := s.duplicates.Submit(ctx, "DUPLICATE", map[string]any{
				"account_id":    cmd.AccountID,
				"amount_minor":  cmd.AmountMinor,
				"date":          cmd.Date,
				"type":          cmd.Type,
				"description":   cmd.Description,
				"collides_with": m.ID,
			}, nil)
			if parkErr != nil {
				// The collision is real either way: report the conflict and
				// let the queue retry. A parked-review outage must not turn
				// a duplicate into a fabricated success.
				log.Info(fmt.Sprintf("[MOVEMENT CREATE] failed to park duplicate: %v", parkErr))
				return domain.ManualMovementOutcome{Duplicate: true}, nil
			}
			return domain.ManualMovementOutcome{Duplicate: true, ParkedReviewID: itemID}, nil
		}
	}

	now := domain.NowRFC3339()
	feID := domain.NewID("fe_manual")
	var catID *string
	if cmd.CategoryID != "" {
		catID = &cmd.CategoryID
	}
	// ToAccountID never reaches this path with a value: a TRANSFER with a
	// destination is written by TransferService (two legs, one relation)
	// before CreateManual is reached; this path records the income or
	// expense side exactly as the manual form submits it.
	fe := domain.FinancialEvent{
		ID:                 feID,
		EventType:          cmd.Type,
		EventDate:          &cmd.Date,
		Status:             "ACTIVE",
		VerificationStatus: "UNVERIFIED",
		Description:        &cmd.Description,
		CategoryID:         catID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	mvID := domain.NewID("mv_manual")
	mv := domain.Movement{
		ID:               mvID,
		FinancialEventID: feID,
		AccountID:        cmd.AccountID,
		MovementDate:     cmd.Date,
		AmountMinor:      cmd.AmountMinor,
		Description:      &cmd.Description,
		Status:           "ACTIVE",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	created, err := s.CreateWithEvent(ctx, fe, mv, nil)
	if err != nil {
		return domain.ManualMovementOutcome{}, err
	}
	log.Info(fmt.Sprintf("[MOVEMENT CREATE] Created movement %s", created.ID))
	return domain.ManualMovementOutcome{ID: created.ID}, nil
}
