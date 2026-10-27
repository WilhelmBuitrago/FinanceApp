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

// Create inserts a movement. Referential integrity is enforced before the
// insert (the SQLite schema declares no FKs on movements). Callers that need
// movement + components atomically use app.MovementService, which orchestrates
// this in a transaction.
func (r *MovementRepository) Create(ctx context.Context, mv domain.Movement) (domain.Movement, error) {
	if r.db.Querier == nil {
		return domain.Movement{}, ErrUnavailable
	}
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
	return mapDBMovementToDomain(dbMv), nil
}

// CreateComponent inserts one component of a movement.
func (r *MovementRepository) CreateComponent(ctx context.Context, comp domain.MovementComponent) (domain.MovementComponent, error) {
	if r.db.Querier == nil {
		return domain.MovementComponent{}, ErrUnavailable
	}
	cparams := db.CreateMovementComponentParams{
		ID:            comp.ID,
		MovementID:    comp.MovementID,
		ComponentType: comp.ComponentType,
		AmountMinor:   comp.AmountMinor,
		Description:   toNullString(comp.Description),
		MetadataJson:  toNullString(comp.MetadataJson),
		CreatedAt:     comp.CreatedAt,
		UpdatedAt:     comp.UpdatedAt,
	}
	dbComp, err := r.db.Querier.CreateMovementComponent(ctx, cparams)
	if err != nil {
		return domain.MovementComponent{}, translateError(err)
	}
	return mapDBMovementComponentToDomain(dbComp), nil
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

func (r *MovementRepository) List(ctx context.Context, limit int) ([]domain.Movement, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbMvs, err := r.db.Querier.ListMovementsAll(ctx, int64(limit))
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Movement, 0, len(dbMvs))
	for _, m := range dbMvs {
		out = append(out, mapDBMovementToDomain(m))
	}
	return out, nil
}

// ListWithCategory returns every movement joined with the category of its
// financial event: the listings need both, and one join replaces the two
// per-row lookups they used to make.
func (r *MovementRepository) ListWithCategory(ctx context.Context, limit int) ([]domain.MovementWithCategory, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	rows, err := r.db.Querier.ListMovementsWithEventAndCategory(ctx, int64(limit))
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.MovementWithCategory, 0, len(rows))
	for _, m := range rows {
		out = append(out, mapDBMovementWithCategoryToDomain(m))
	}
	return out, nil
}

// ListByAccountWithCategory is ListWithCategory scoped to one account.
func (r *MovementRepository) ListByAccountWithCategory(ctx context.Context, accountID string, limit int) ([]domain.MovementWithCategory, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	rows, err := r.db.Querier.ListMovementsByAccountWithEventAndCategory(ctx, db.ListMovementsByAccountWithEventAndCategoryParams{
		AccountID: accountID,
		Limit:     int64(limit),
	})
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.MovementWithCategory, 0, len(rows))
	for _, m := range rows {
		out = append(out, mapDBMovementByAccountWithCategoryToDomain(m))
	}
	return out, nil
}

// ListByCategory returns movements whose financial event carries the given
// category: used when a category is removed to park them for human review.
func (r *MovementRepository) ListByCategory(ctx context.Context, categoryID string) ([]domain.Movement, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbMvs, err := r.db.Querier.ListMovementsByCategory(ctx, sql.NullString{String: categoryID, Valid: true})
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Movement, 0, len(dbMvs))
	for _, m := range dbMvs {
		out = append(out, mapDBMovementToDomain(m))
	}
	return out, nil
}

// FindByNaturalKey returns the movements that collide with a manual entry
// on account, date and amount, with a description that is empty on either
// side. One indexed query replaces loading the entire movements table.
func (r *MovementRepository) FindByNaturalKey(ctx context.Context, accountID, date string, amountMinor int64, description string) ([]domain.Movement, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	rows, err := r.db.Querier.FindMovementsByNaturalKey(ctx, db.FindMovementsByNaturalKeyParams{
		AccountID:          accountID,
		MovementDate:       date,
		AmountMinor:        amountMinor,
		RequestDescription: toNullString(&description),
	})
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Movement, 0, len(rows))
	for _, m := range rows {
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

func movementWithCategoryToDomain(mv db.Movement, eventType string, categoryID, categoryName, counterpartID, counterpartName sql.NullString) domain.MovementWithCategory {
	return domain.MovementWithCategory{
		Movement:               mapDBMovementToDomain(mv),
		CategoryID:             fromNullStringValue(categoryID),
		CategoryName:           fromNullStringValue(categoryName),
		EventType:              eventType,
		CounterpartAccountName: fromNullStringValue(counterpartName),
	}
}

func mapDBMovementWithCategoryToDomain(m db.ListMovementsWithEventAndCategoryRow) domain.MovementWithCategory {
	return movementWithCategoryToDomain(db.Movement{
		ID:               m.ID,
		FinancialEventID: m.FinancialEventID,
		AccountID:        m.AccountID,
		SourceEntryID:    m.SourceEntryID,
		MovementDate:     m.MovementDate,
		AmountMinor:      m.AmountMinor,
		Description:      m.Description,
		Status:           m.Status,
		MetadataJson:     m.MetadataJson,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}, m.EventType, m.CategoryID, m.CategoryName, m.CounterpartAccountID, m.CounterpartAccountName)
}

func mapDBMovementByAccountWithCategoryToDomain(m db.ListMovementsByAccountWithEventAndCategoryRow) domain.MovementWithCategory {
	return movementWithCategoryToDomain(db.Movement{
		ID:               m.ID,
		FinancialEventID: m.FinancialEventID,
		AccountID:        m.AccountID,
		SourceEntryID:    m.SourceEntryID,
		MovementDate:     m.MovementDate,
		AmountMinor:      m.AmountMinor,
		Description:      m.Description,
		Status:           m.Status,
		MetadataJson:     m.MetadataJson,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}, m.EventType, m.CategoryID, m.CategoryName, m.CounterpartAccountID, m.CounterpartAccountName)
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
