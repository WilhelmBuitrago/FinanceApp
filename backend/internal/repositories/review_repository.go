package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"financeapp/backend/internal/db"
)

// ReviewEntry is a movement the parser parked for human verification: the
// raw is preserved so the user checks it against the PDF.
type ReviewEntry struct {
	ID             string
	DocumentID     string
	RawDescription *string
	RawDate        *string
	PageNumber     *int64
	Reason         string
	CreatedAt      string
}

type ReviewRepository struct {
	db *DB
}

func NewReviewRepository(db *DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) ListEntries(ctx context.Context) ([]ReviewEntry, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	rows, err := r.db.Querier.ListReviewEntries(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]ReviewEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, ReviewEntry{
			ID:             row.ID,
			DocumentID:     row.DocumentID,
			RawDescription: fromNullString(row.RawDescription),
			RawDate:        fromNullString(row.RawDate),
			PageNumber:     fromNullInt64(row.PageNumber),
			Reason:         reasonFromMetadata(row.MetadataJson),
			CreatedAt:      row.CreatedAt,
		})
	}
	return out, nil
}

// Resolve parks the human correction on the source entry. Zero rows affected
// means the entry was already resolved (or never existed): ErrConflict, so a
// double submission cannot create the movement twice.
func (r *ReviewRepository) Resolve(ctx context.Context, id string, correctedDescription *string, correctedAmountMinor *int64, correctedDate *string, updatedAt string) error {
	if r.db.Querier == nil {
		return ErrUnavailable
	}
	affected, err := r.db.Querier.ResolveSourceEntry(ctx, db.ResolveSourceEntryParams{
		ID:                   id,
		CorrectedDescription: toNullString(correctedDescription),
		CorrectedAmountMinor: toNullInt64(correctedAmountMinor),
		CorrectedDate:        toNullString(correctedDate),
		UpdatedAt:            updatedAt,
	})
	if err != nil {
		return translateError(err)
	}
	if affected == 0 {
		return ErrConflict
	}
	return nil
}

func reasonFromMetadata(meta sql.NullString) string {
	if !meta.Valid || meta.String == "" {
		return ""
	}
	var parsed struct {
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(meta.String), &parsed); err != nil {
		return ""
	}
	return parsed.Reason
}
