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
		SourceReference:       toInterface(se.SourceReference),
		RawDescription:        toInterface(se.RawDescription),
		RawAmountMinor:        toInterfaceInt64(se.RawAmountMinor),
		RawDate:               toInterface(se.RawDate),
		NormalizedDescription: toInterface(se.NormalizedDescription),
		NormalizedAmountMinor: toInterfaceInt64(se.NormalizedAmountMinor),
		NormalizedDate:        toInterface(se.NormalizedDate),
		PageNumber:            toInterfaceInt64(se.PageNumber),
		RowNumber:             toInterfaceInt64(se.RowNumber),
		Section:               toInterface(se.Section),
		ExtractionID:          toInterface(se.ExtractionID),
		MetadataJson:          toInterface(se.MetadataJson),
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

func (r *SourceEntryRepository) Update(ctx context.Context, se domain.SourceEntry) (domain.SourceEntry, error) {
	return domain.SourceEntry{}, ErrInvalid
}

func mapDBSourceEntryToDomain(s db.SourceEntry) domain.SourceEntry {
	int64Ptr := func(v interface{}) *int64 {
		if v == nil {
			return nil
		}
		i := v.(int64)
		return &i
	}
	strPtr := func(v interface{}) *string {
		if v == nil {
			return nil
		}
		sv := v.(string)
		return &sv
	}
	return domain.SourceEntry{
		ID:                    s.ID,
		DocumentID:            s.DocumentID,
		SourceReference:       strPtr(s.SourceReference),
		RawDescription:        strPtr(s.RawDescription),
		RawAmountMinor:        int64Ptr(s.RawAmountMinor),
		RawDate:               strPtr(s.RawDate),
		NormalizedDescription: strPtr(s.NormalizedDescription),
		NormalizedAmountMinor: int64Ptr(s.NormalizedAmountMinor),
		NormalizedDate:        strPtr(s.NormalizedDate),
		PageNumber:            int64Ptr(s.PageNumber),
		RowNumber:             int64Ptr(s.RowNumber),
		Section:               strPtr(s.Section),
		ExtractionID:          strPtr(s.ExtractionID),
		MetadataJson:          strPtr(s.MetadataJson),
		CreatedAt:             s.CreatedAt,
		UpdatedAt:             s.UpdatedAt,
	}
}
