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
		EventDate:          toInterface(fe.EventDate),
		Status:             fe.Status,
		VerificationStatus: fe.VerificationStatus,
		Description:        toInterface(fe.Description),
		CategoryID:         toInterface(fe.CategoryID),
		Confidence:         toInterfaceFloat64(fe.Confidence),
		MetadataJson:       toInterface(fe.MetadataJson),
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

func (r *FinancialEventRepository) Update(ctx context.Context, fe domain.FinancialEvent) (domain.FinancialEvent, error) {
	return domain.FinancialEvent{}, ErrInvalid
}

func mapDBFinancialEventToDomain(f db.FinancialEvent) domain.FinancialEvent {
	strPtr := func(v interface{}) *string {
		if v == nil {
			return nil
		}
		s := v.(string)
		return &s
	}
	floatPtr := func(v interface{}) *float64 {
		if v == nil {
			return nil
		}
		fv := v.(float64)
		return &fv
	}
	return domain.FinancialEvent{
		ID:                 f.ID,
		EventType:          f.EventType,
		EventDate:          strPtr(f.EventDate),
		Status:             f.Status,
		VerificationStatus: f.VerificationStatus,
		Description:        strPtr(f.Description),
		CategoryID:         strPtr(f.CategoryID),
		Confidence:         floatPtr(f.Confidence),
		MetadataJson:       strPtr(f.MetadataJson),
		CreatedAt:          f.CreatedAt,
		UpdatedAt:          f.UpdatedAt,
	}
}
