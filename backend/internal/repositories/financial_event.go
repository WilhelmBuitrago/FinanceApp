package repositories

import (
	"context"
	"database/sql"
	"errors"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
)

type FinancialEventRepository struct {
	db *DB
}

func NewFinancialEventRepository(db *DB) *FinancialEventRepository {
	return &FinancialEventRepository{db: db}
}

func (r *FinancialEventRepository) Create(ctx context.Context, fe domain.FinancialEvent) (domain.FinancialEvent, error) {
	if r.db.Querier == nil {
		return domain.FinancialEvent{}, ErrUnavailable
	}
	params := db.CreateFinancialEventParams{
		ID:                 fe.ID,
		EventType:          fe.EventType,
		EventDate:          toNullString(fe.EventDate),
		Status:             fe.Status,
		VerificationStatus: fe.VerificationStatus,
		Description:        toNullString(fe.Description),
		CategoryID:         toNullString(fe.CategoryID),
		Confidence:         toNullFloat64(fe.Confidence),
		MetadataJson:       toNullString(fe.MetadataJson),
		CreatedAt:          fe.CreatedAt,
		UpdatedAt:          fe.UpdatedAt,
	}
	dbFE, err := r.db.Querier.CreateFinancialEvent(ctx, params)
	if err != nil {
		return domain.FinancialEvent{}, translateError(err)
	}
	return mapDBFinancialEventToDomain(dbFE), nil
}

func (r *FinancialEventRepository) GetByID(ctx context.Context, id string) (domain.FinancialEvent, error) {
	if r.db.Querier == nil {
		return domain.FinancialEvent{}, ErrUnavailable
	}
	dbFE, err := r.db.Querier.GetFinancialEvent(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FinancialEvent{}, ErrNotFound
		}
		return domain.FinancialEvent{}, translateError(err)
	}
	return mapDBFinancialEventToDomain(dbFE), nil
}

func (r *FinancialEventRepository) ListByType(ctx context.Context, eventType string) ([]domain.FinancialEvent, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbFEs, err := r.db.Querier.ListFinancialEventsByType(ctx, eventType)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.FinancialEvent, 0, len(dbFEs))
	for _, f := range dbFEs {
		out = append(out, mapDBFinancialEventToDomain(f))
	}
	return out, nil
}

func mapDBFinancialEventToDomain(f db.FinancialEvent) domain.FinancialEvent {
	return domain.FinancialEvent{
		ID:                 f.ID,
		EventType:          f.EventType,
		EventDate:          fromNullString(f.EventDate),
		Status:             f.Status,
		VerificationStatus: f.VerificationStatus,
		Description:        fromNullString(f.Description),
		CategoryID:         fromNullString(f.CategoryID),
		Confidence:         fromNullFloat64(f.Confidence),
		MetadataJson:       fromNullString(f.MetadataJson),
		CreatedAt:          f.CreatedAt,
		UpdatedAt:          f.UpdatedAt,
	}
}
