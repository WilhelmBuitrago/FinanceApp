package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

// Review kinds parked in the unified queue.
const (
	reviewKindDuplicate      = "DUPLICATE"
	reviewKindCategory       = "CATEGORY"
	reviewKindExtraction     = "EXTRACTION"
	reviewKindDocumentPeriod = "DOCUMENT_PERIOD"
)

// categoryPayload is what a CATEGORY item carries: a movement whose category
// disappeared and the category that was there.
type categoryPayload struct {
	MovementID   string  `json:"movement_id"`
	CategoryID   *string `json:"category_id"`
	CategoryName string  `json:"category_name"`
}

// ReviewCorrection is the human's decision for one queue item: for
// DUPLICATE/EXTRACTION it is the movement to create; for CATEGORY it only
// needs CategoryID (nil = leave the movement uncategorized).
type ReviewCorrection struct {
	Description string
	AmountMinor int64
	Date        string
	Type        string
	AccountID   string
	CategoryID  *string
}

// ReviewService owns the unified human-review queue. Triggers submit items
// (Submit) and a human resolves or discards them (Resolve/Discard).
type ReviewService struct {
	db *sql.DB
}

func NewReviewService(db *sql.DB) *ReviewService {
	return &ReviewService{db: db}
}

// Submit parks one item of the given kind with its JSON payload.
func (s *ReviewService) Submit(ctx context.Context, kind string, payload any, sourceEntryID *string) (string, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal review payload: %w", err)
	}
	id := uuid.NewString()
	repo := repositories.NewReviewItemsRepository(&repositories.DB{Querier: db.New(s.db)})
	if err := repo.Create(ctx, id, kind, strPtr(string(payloadJSON)), sourceEntryID, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return "", fmt.Errorf("park review item: %w", err)
	}
	return id, nil
}

// ListPending returns the queue for the UI.
func (s *ReviewService) ListPending(ctx context.Context) ([]repositories.ReviewItemRecord, error) {
	repo := repositories.NewReviewItemsRepository(&repositories.DB{Querier: db.New(s.db)})
	return repo.ListPending(ctx)
}

// Resolve applies the human's decision: marks the item resolved first (a
// second submission fails before any money moves), then performs the action
// the kind demands, all in one transaction. The correction fields are the
// movement to create (DUPLICATE/EXTRACTION) or the category to apply
// (CATEGORY, via categoryID).
func (s *ReviewService) Resolve(ctx context.Context, itemID string, description string, amountMinor int64, date string, eventType string, accountID string, categoryID *string) (created domain.Movement, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Movement{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC().Format(time.RFC3339)
	txQuerier := db.New(tx)
	itemsRepo := repositories.NewReviewItemsRepository(&repositories.DB{Querier: txQuerier})

	item, err := itemsRepo.Get(ctx, itemID)
	if err != nil {
		return domain.Movement{}, err
	}
	if err = itemsRepo.Resolve(ctx, itemID, now); err != nil {
		return domain.Movement{}, err
	}

	correction := ReviewCorrection{
		Description: description,
		AmountMinor: amountMinor,
		Date:        date,
		Type:        eventType,
		AccountID:   accountID,
		CategoryID:  categoryID,
	}

	switch item.Kind {
	case reviewKindCategory:
		if err = applyCategoryDecision(ctx, txQuerier, item, correction); err != nil {
			return domain.Movement{}, fmt.Errorf("apply category decision: %w", err)
		}
	case reviewKindExtraction:
		if err = markSourceEntryResolved(ctx, txQuerier, item, correction); err != nil {
			return domain.Movement{}, fmt.Errorf("mark extraction resolved: %w", err)
		}
		if created, err = createMovementFromCorrection(ctx, txQuerier, correction); err != nil {
			return domain.Movement{}, err
		}
	case reviewKindDuplicate:
		if created, err = createMovementFromCorrection(ctx, txQuerier, correction); err != nil {
			return domain.Movement{}, err
		}
	default:
		return domain.Movement{}, fmt.Errorf("unknown review kind: %s", item.Kind)
	}

	if err = tx.Commit(); err != nil {
		return domain.Movement{}, fmt.Errorf("commit: %w", err)
	}
	return created, nil
}

// Discard parks an item as discarded: the human decided it is not real.
func (s *ReviewService) Discard(ctx context.Context, itemID string) error {
	repo := repositories.NewReviewItemsRepository(&repositories.DB{Querier: db.New(s.db)})
	return repo.Discard(ctx, itemID, time.Now().UTC().Format(time.RFC3339))
}

func ptrText(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// toNullString converts an optional string into a nullable text column value.
func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func createMovementFromCorrection(ctx context.Context, q db.Querier, correction ReviewCorrection) (domain.Movement, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	fe := domain.FinancialEvent{
		ID:                 uuid.NewString(),
		EventType:          correction.Type,
		EventDate:          &correction.Date,
		Status:             "ACTIVE",
		VerificationStatus: "VERIFIED", // a human checked this one
		Description:        &correction.Description,
		CategoryID:         correction.CategoryID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	mv := domain.Movement{
		ID:               uuid.NewString(),
		FinancialEventID: fe.ID,
		AccountID:        correction.AccountID,
		MovementDate:     correction.Date,
		AmountMinor:      correction.AmountMinor,
		Description:      &correction.Description,
		Status:           "ACTIVE",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	return writeEventAndMovement(ctx, q, fe, mv, nil)
}

func applyCategoryDecision(ctx context.Context, q db.Querier, item repositories.ReviewItemRecord, correction ReviewCorrection) error {
	var payload categoryPayload
	if err := json.Unmarshal([]byte(ptrText(item.PayloadJson)), &payload); err != nil {
		return fmt.Errorf("read category payload: %w", err)
	}
	movementRepo := repositories.NewMovementRepository(&repositories.DB{Querier: q})
	mv, err := movementRepo.GetByID(ctx, payload.MovementID)
	if err != nil {
		return fmt.Errorf("movement not found: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	affected, err := q.UpdateFinancialEventCategory(ctx, db.UpdateFinancialEventCategoryParams{
		ID:         mv.FinancialEventID,
		CategoryID: toNullString(correction.CategoryID),
		UpdatedAt:  now,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("financial event %s not found", mv.FinancialEventID)
	}
	return nil
}

func markSourceEntryResolved(ctx context.Context, q db.Querier, item repositories.ReviewItemRecord, correction ReviewCorrection) error {
	if item.SourceEntryID == nil {
		return nil
	}
	repo := repositories.NewReviewRepository(&repositories.DB{Querier: q})
	return repo.Resolve(ctx, *item.SourceEntryID, &correction.Description, &correction.AmountMinor, &correction.Date, time.Now().UTC().Format(time.RFC3339))
}
