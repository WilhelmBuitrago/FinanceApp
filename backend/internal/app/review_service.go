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
	"financeapp/backend/internal/log"
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
func (s *ReviewService) ListPending(ctx context.Context, limit int) ([]repositories.ReviewItemRecord, error) {
	repo := repositories.NewReviewItemsRepository(&repositories.DB{Querier: db.New(s.db)})
	return repo.ListPending(ctx, limit)
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

// Kind returns the queue item's kind so the HTTP layer can validate the
// correction per kind before resolving: a CATEGORY item only needs the
// category; the movement-creating kinds need the full correction.
func (s *ReviewService) Kind(ctx context.Context, itemID string) (string, error) {
	repo := repositories.NewReviewItemsRepository(&repositories.DB{Querier: db.New(s.db)})
	item, err := repo.Get(ctx, itemID)
	if err != nil {
		return "", err
	}
	return item.Kind, nil
}

// Discard parks an item as discarded: the human decided it is not real.
func (s *ReviewService) Discard(ctx context.Context, itemID string) error {
	repo := repositories.NewReviewItemsRepository(&repositories.DB{Querier: db.New(s.db)})
	return repo.Discard(ctx, itemID, time.Now().UTC().Format(time.RFC3339))
}

// documentPeriodPayload is what a DOCUMENT_PERIOD item carries: the raw
// period the processor could not read, plus the ids the resolution needs.
type documentPeriodPayload struct {
	RawPeriod  string `json:"raw_period"`
	Reason     string `json:"reason"`
	DocumentID string `json:"document_id"`
	AccountID  string `json:"account_id"`
}

// ResolvePeriod records the human-confirmed month/year for a parked
// statement period and derives against it the parked movement dates of
// THAT document: the day comes from the movement's raw date, the month must
// be the confirmed one, and the day must exist in that month of that year.
// The movements that fit are imported in the same transaction and their
// entries and queue items are marked resolved; the ones that do not fit
// stay parked for individual human review. Already-imported movements are
// never touched.
func (s *ReviewService) ResolvePeriod(ctx context.Context, itemID string, month, year int) error {
	if month < 1 || month > 12 {
		return fmt.Errorf("month out of range: %d", month)
	}
	if year < 2000 || year > 2100 {
		return fmt.Errorf("year out of range: %d", year)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
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
		return err
	}
	if item.Kind != reviewKindDocumentPeriod {
		return fmt.Errorf("review item %s is not a document period item: kind %s", itemID, item.Kind)
	}
	// The item closes before any money moves: a second submission fails.
	if err = itemsRepo.Resolve(ctx, itemID, now); err != nil {
		return err
	}

	var payload documentPeriodPayload
	if err := json.Unmarshal([]byte(ptrText(item.PayloadJson)), &payload); err != nil {
		return fmt.Errorf("read document period payload: %w", err)
	}
	if payload.DocumentID == "" || payload.AccountID == "" {
		return fmt.Errorf("document period payload missing document_id or account_id")
	}

	start, end := PeriodRangeISO(year, month)
	if _, err = txQuerier.UpdateDocumentPeriod(ctx, db.UpdateDocumentPeriodParams{
		ID:          payload.DocumentID,
		PeriodStart: toNullString(&start),
		PeriodEnd:   toNullString(&end),
		UpdatedAt:   now,
	}); err != nil {
		return fmt.Errorf("update document period: %w", err)
	}

	entries, err := txQuerier.ListReviewRequiredByDocument(ctx, payload.DocumentID)
	if err != nil {
		return fmt.Errorf("list parked entries: %w", err)
	}
	derived := 0
	for _, entry := range entries {
		iso, fits := DeriveMovementDate(entry.RawDate.String, year, month)
		if !fits {
			continue
		}
		var amount int64
		if entry.NormalizedAmountMinor.Valid {
			amount = entry.NormalizedAmountMinor.Int64
		}
		description := entry.NormalizedDescription.String
		if description == "" {
			description = entry.RawDescription.String
		}
		fe := domain.FinancialEvent{
			ID:                 uuid.NewString(),
			EventType:          classifyEventType(amount),
			EventDate:          &iso,
			Status:             "ACTIVE",
			VerificationStatus: "VERIFIED",
			Description:        &description,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		mv := domain.Movement{
			ID:               uuid.NewString(),
			FinancialEventID: fe.ID,
			AccountID:        payload.AccountID,
			MovementDate:     iso,
			AmountMinor:      amount,
			Description:      &description,
			Status:           "ACTIVE",
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if _, err = writeEventAndMovement(ctx, txQuerier, fe, mv, nil); err != nil {
			return fmt.Errorf("create derived movement: %w", err)
		}
		if err = txQuerier.ResolveSourceEntryDerivedDate(ctx, db.ResolveSourceEntryDerivedDateParams{
			ID:                    entry.ID,
			NormalizedDate:        toNullString(&iso),
			NormalizedDescription: toNullString(&description),
			NormalizedAmountMinor: sql.NullInt64{Int64: amount, Valid: true},
			UpdatedAt:             now,
		}); err != nil {
			return fmt.Errorf("mark derived entry: %w", err)
		}
		if _, err = txQuerier.ResolveReviewItemsBySourceEntry(ctx, db.ResolveReviewItemsBySourceEntryParams{
			SourceEntryID: toNullString(&entry.ID),
			ResolvedAt:    toNullString(&now),
		}); err != nil {
			return fmt.Errorf("close parked review item: %w", err)
		}
		derived++
	}
	log.Info(fmt.Sprintf(
		"Review period resolved document=%s period=%04d-%02d derived=%d parked_left=%d",
		payload.DocumentID, year, month, derived, len(entries)-derived))

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
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
	repo := repositories.NewSourceEntryRepository(&repositories.DB{Querier: q})
	return repo.Resolve(ctx, *item.SourceEntryID, &correction.Description, &correction.AmountMinor, &correction.Date, time.Now().UTC().Format(time.RFC3339))
}
