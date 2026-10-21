package repositories

import (
	"context"
	"database/sql"
	"errors"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
)

type MovementRepository struct {
	db *DB
}

func NewMovementRepository(db *DB) *MovementRepository {
	return &MovementRepository{db: db}
}

func (r *MovementRepository) Create(ctx context.Context, mv domain.Movement, components []domain.MovementComponent) (domain.Movement, error) {
	if r.db.Querier == nil {
		return domain.Movement{}, ErrUnavailable
	}
	// The SQLite schema does not declare foreign keys on movements, so the
	// repository enforces referential integrity before inserting.
	if _, err := r.db.Querier.GetAccount(ctx, mv.AccountID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Movement{}, ErrConstraint
		}
		return domain.Movement{}, translateError(err)
	}
	if _, err := r.db.Querier.GetFinancialEvent(ctx, mv.FinancialEventID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Movement{}, ErrConstraint
		}
		return domain.Movement{}, translateError(err)
	}
	// In real implementation, use transaction. Here we simulate with Querier.
	params := db.CreateMovementParams{
		ID:               mv.ID,
		FinancialEventID: mv.FinancialEventID,
		AccountID:        mv.AccountID,
		SourceEntryID:    toNullString(mv.SourceEntryID),
		MovementDate:     mv.MovementDate,
		AmountMinor:      mv.AmountMinor,
		Description:      toNullString(mv.Description),
		Status:           mv.Status,
		MetadataJson:     toNullString(mv.MetadataJson),
		CreatedAt:        mv.CreatedAt,
		UpdatedAt:        mv.UpdatedAt,
	}
	dbMv, err := r.db.Querier.CreateMovement(ctx, params)
	if err != nil {
		return domain.Movement{}, translateError(err)
	}
	// Create components
	for _, comp := range components {
		cparams := db.CreateMovementComponentParams{
			ID:            comp.ID,
			MovementID:    dbMv.ID,
			ComponentType: comp.ComponentType,
			AmountMinor:   comp.AmountMinor,
			Description:   toNullString(comp.Description),
			MetadataJson:  toNullString(comp.MetadataJson),
			CreatedAt:     comp.CreatedAt,
			UpdatedAt:     comp.UpdatedAt,
		}
		_, err := r.db.Querier.CreateMovementComponent(ctx, cparams)
		if err != nil {
			return domain.Movement{}, translateError(err)
		}
	}
	return mapDBMovementToDomain(dbMv), nil
}

func (r *MovementRepository) GetByID(ctx context.Context, id string) (domain.Movement, error) {
	if r.db.Querier == nil {
		return domain.Movement{}, ErrUnavailable
	}
	dbMv, err := r.db.Querier.GetMovement(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Movement{}, ErrNotFound
		}
		return domain.Movement{}, translateError(err)
	}
	return mapDBMovementToDomain(dbMv), nil
}

func (r *MovementRepository) ListByAccount(ctx context.Context, accountID string) ([]domain.Movement, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbMvs, err := r.db.Querier.ListMovementsByAccount(ctx, accountID)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Movement, 0, len(dbMvs))
	for _, m := range dbMvs {
		out = append(out, mapDBMovementToDomain(m))
	}
	return out, nil
}

func (r *MovementRepository) ListComponents(ctx context.Context, movementID string) ([]domain.MovementComponent, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbComps, err := r.db.Querier.ListComponentsByMovement(ctx, movementID)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.MovementComponent, 0, len(dbComps))
	for _, c := range dbComps {
		out = append(out, mapDBMovementComponentToDomain(c))
	}
	return out, nil
}

func (r *MovementRepository) List(ctx context.Context) ([]domain.Movement, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbMvs, err := r.db.Querier.ListMovementsAll(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Movement, 0, len(dbMvs))
	for _, m := range dbMvs {
		out = append(out, mapDBMovementToDomain(m))
	}
	return out, nil
}

func (r *MovementRepository) Delete(ctx context.Context, id string) error {
	if r.db.Querier == nil {
		return ErrUnavailable
	}
	if err := r.db.Querier.DeleteMovement(ctx, id); err != nil {
		return translateError(err)
	}
	return nil
}

func (r *MovementRepository) Update(ctx context.Context, mv domain.Movement) (domain.Movement, error) {
	return domain.Movement{}, ErrInvalid
}

func mapDBMovementToDomain(m db.Movement) domain.Movement {
	return domain.Movement{
		ID:               m.ID,
		FinancialEventID: m.FinancialEventID,
		AccountID:        m.AccountID,
		SourceEntryID:    fromNullString(m.SourceEntryID),
		MovementDate:     m.MovementDate,
		AmountMinor:      m.AmountMinor,
		Description:      fromNullString(m.Description),
		Status:           m.Status,
		MetadataJson:     fromNullString(m.MetadataJson),
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func mapDBMovementComponentToDomain(c db.MovementComponent) domain.MovementComponent {
	return domain.MovementComponent{
		ID:            c.ID,
		MovementID:    c.MovementID,
		ComponentType: c.ComponentType,
		AmountMinor:   c.AmountMinor,
		Description:   fromNullString(c.Description),
		MetadataJson:  fromNullString(c.MetadataJson),
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}
