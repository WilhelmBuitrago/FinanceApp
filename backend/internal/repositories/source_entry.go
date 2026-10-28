package repositories

import (
	"context"
	"database/sql"
	"errors"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
)

type SourceEntryRepository struct {
	db *DB
}

func NewSourceEntryRepository(db *DB) *SourceEntryRepository {
	return &SourceEntryRepository{db: db}
}

func (r *SourceEntryRepository) Create(ctx context.Context, se domain.SourceEntry) (domain.SourceEntry, error) {
	if r.db.Querier == nil {
		return domain.SourceEntry{}, ErrUnavailable
	}
	params := db.CreateSourceEntryParams{
		ID:                    se.ID,
		DocumentID:            se.DocumentID,
		SourceReference:       toNullString(se.SourceReference),
		RawDescription:        toNullString(se.RawDescription),
		RawAmountMinor:        toNullInt64(se.RawAmountMinor),
		RawDate:               toNullString(se.RawDate),
		NormalizedDescription: toNullString(se.NormalizedDescription),
		NormalizedAmountMinor: toNullInt64(se.NormalizedAmountMinor),
		NormalizedDate:        toNullString(se.NormalizedDate),
		PageNumber:            toNullInt64(se.PageNumber),
		RowNumber:             toNullInt64(se.RowNumber),
		Section:               toNullString(se.Section),
		ExtractionID:          toNullString(se.ExtractionID),
		ReviewStatus:          se.ReviewStatus,
		SourceKey:             toNullString(se.SourceKey),
		MetadataJson:          toNullString(se.MetadataJson),
		CreatedAt:             se.CreatedAt,
		UpdatedAt:             se.UpdatedAt,
	}
	dbSE, err := r.db.Querier.CreateSourceEntry(ctx, params)
	if err != nil {
		return domain.SourceEntry{}, translateError(err)
	}
	return mapDBSourceEntryToDomain(dbSE), nil
}

func (r *SourceEntryRepository) GetByID(ctx context.Context, id string) (domain.SourceEntry, error) {
	if r.db.Querier == nil {
		return domain.SourceEntry{}, ErrUnavailable
	}
	dbSE, err := r.db.Querier.GetSourceEntry(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SourceEntry{}, ErrNotFound
		}
		return domain.SourceEntry{}, translateError(err)
	}
	return mapDBSourceEntryToDomain(dbSE), nil
}

func (r *SourceEntryRepository) ListByDocument(ctx context.Context, documentID string) ([]domain.SourceEntry, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbSEs, err := r.db.Querier.ListSourceEntriesByDocument(ctx, documentID)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.SourceEntry, 0, len(dbSEs))
	for _, s := range dbSEs {
		// sqlc unifies the row and the model when the query selects every
		// column: the conversion keeps one mapper for both shapes.
		out = append(out, mapDBSourceEntryToDomain(db.SourceEntry(s)))
	}
	return out, nil
}

// Resolve parks the human correction on the source entry. Zero rows
// affected means the entry is no longer parked: ErrConflict, so a second
// submission cannot apply the correction twice.
func (r *SourceEntryRepository) Resolve(ctx context.Context, id string, correctedDescription *string, correctedAmountMinor *int64, correctedDate *string, updatedAt string) error {
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

func mapDBSourceEntryToDomain(s db.SourceEntry) domain.SourceEntry {
	return domain.SourceEntry{
		ID:                    s.ID,
		DocumentID:            s.DocumentID,
		SourceReference:       fromNullString(s.SourceReference),
		RawDescription:        fromNullString(s.RawDescription),
		RawAmountMinor:        fromNullInt64(s.RawAmountMinor),
		RawDate:               fromNullString(s.RawDate),
		NormalizedDescription: fromNullString(s.NormalizedDescription),
		NormalizedAmountMinor: fromNullInt64(s.NormalizedAmountMinor),
		NormalizedDate:        fromNullString(s.NormalizedDate),
		PageNumber:            fromNullInt64(s.PageNumber),
		RowNumber:             fromNullInt64(s.RowNumber),
		Section:               fromNullString(s.Section),
		ExtractionID:          fromNullString(s.ExtractionID),
		ReviewStatus:          s.ReviewStatus,
		SourceKey:             fromNullString(s.SourceKey),
		CorrectedDescription:  fromNullString(s.CorrectedDescription),
		CorrectedAmountMinor:  fromNullInt64(s.CorrectedAmountMinor),
		CorrectedDate:         fromNullString(s.CorrectedDate),
		MetadataJson:          fromNullString(s.MetadataJson),
		CreatedAt:             s.CreatedAt,
		UpdatedAt:             s.UpdatedAt,
	}
}
