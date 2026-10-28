package repositories

import (
	"context"
	"database/sql"
	"errors"

	"financeapp/backend/internal/db"
)

// ReviewItemRecord is one entry of the unified human-review queue. Kind is
// DUPLICATE, CATEGORY or EXTRACTION; the payload carries what the UI shows
// and what the resolution needs.
type ReviewItemRecord struct {
	ID            string
	Kind          string
	PayloadJson   *string
	SourceEntryID *string
	Status        string
	CreatedAt     string
	ResolvedAt    *string
}

type ReviewItemsRepository struct {
	db *DB
}

func NewReviewItemsRepository(db *DB) *ReviewItemsRepository {
	return &ReviewItemsRepository{db: db}
}

func (r *ReviewItemsRepository) Create(ctx context.Context, id string, kind string, payloadJson *string, sourceEntryID *string, createdAt string) error {
	if r.db.Querier == nil {
		return ErrUnavailable
	}
	return r.db.Querier.CreateReviewItem(ctx, db.CreateReviewItemParams{
		ID:            id,
		Kind:          kind,
		PayloadJson:   ptrToText(payloadJson),
		SourceEntryID: toNullString(sourceEntryID),
		CreatedAt:     createdAt,
	})
}

func (r *ReviewItemsRepository) ListPending(ctx context.Context, limit int) ([]ReviewItemRecord, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	rows, err := r.db.Querier.ListPendingReviewItems(ctx, int64(limit))
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]ReviewItemRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapDBReviewItem(row))
	}
	return out, nil
}

func (r *ReviewItemsRepository) Get(ctx context.Context, id string) (ReviewItemRecord, error) {
	if r.db.Querier == nil {
		return ReviewItemRecord{}, ErrUnavailable
	}
	row, err := r.db.Querier.GetReviewItem(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ReviewItemRecord{}, ErrNotFound
		}
		return ReviewItemRecord{}, translateError(err)
	}
	return mapDBReviewItem(row), nil
}

// Resolve marks an item resolved. Zero rows affected means it is no longer
// pending: ErrConflict, so the caller cannot act on a stale item twice.
func (r *ReviewItemsRepository) Resolve(ctx context.Context, id string, resolvedAt string) error {
	if r.db.Querier == nil {
		return ErrUnavailable
	}
	affected, err := r.db.Querier.ResolveReviewItem(ctx, db.ResolveReviewItemParams{
		ID:         id,
		ResolvedAt: toNullString(&resolvedAt),
	})
	if err != nil {
		return translateError(err)
	}
	if affected == 0 {
		return ErrConflict
	}
	return nil
}

// Discard parks an item as discarded (the human decided it is not real).
func (r *ReviewItemsRepository) Discard(ctx context.Context, id string, resolvedAt string) error {
	if r.db.Querier == nil {
		return ErrUnavailable
	}
	affected, err := r.db.Querier.DiscardReviewItem(ctx, db.DiscardReviewItemParams{
		ID:         id,
		ResolvedAt: toNullString(&resolvedAt),
	})
	if err != nil {
		return translateError(err)
	}
	if affected == 0 {
		return ErrConflict
	}
	return nil
}

func mapDBReviewItem(row db.ReviewItem) ReviewItemRecord {
	return reviewItemFrom(row.ID, row.Kind, row.PayloadJson, row.SourceEntryID, row.Status, row.CreatedAt, row.ResolvedAt)
}

func reviewItemFrom(id, kind, payloadJson string, sourceEntryID sql.NullString, status, createdAt string, resolvedAt sql.NullString) ReviewItemRecord {
	var payload *string
	if payloadJson != "" {
		payload = &payloadJson
	}
	return ReviewItemRecord{
		ID:            id,
		Kind:          kind,
		PayloadJson:   payload,
		SourceEntryID: fromNullString(sourceEntryID),
		Status:        status,
		CreatedAt:     createdAt,
		ResolvedAt:    fromNullString(resolvedAt),
	}
}

// ptrToText converts an optional payload string into the non-null text column.
func ptrToText(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
