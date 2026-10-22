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
		out = append(out, mapDBSourceEntryToDomain(s))
	}
	return out, nil
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
