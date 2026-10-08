package repositories

import (
	"context"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
)

type RelationRepository struct {
	db *DB
}

func NewRelationRepository(db *DB) *RelationRepository {
	return &RelationRepository{db: db}
}

func (r *RelationRepository) CreateMovementRelation(ctx context.Context, rel domain.MovementRelation) (domain.MovementRelation, error) {
	if r.db.Querier == nil {
		return domain.MovementRelation{}, ErrUnavailable
	}
	params := db.CreateMovementRelationParams{
		ID:           rel.ID,
		MovementIDA:  rel.MovementIDA,
		MovementIDB:  rel.MovementIDB,
		RelationType: rel.RelationType,
		Status:       rel.Status,
		Confidence:   toInterfaceFloat64(rel.Confidence),
		Reason:       toInterface(rel.Reason),
		CreatedAt:    rel.CreatedAt,
		UpdatedAt:    rel.UpdatedAt,
	}
	dbRel, err := r.db.Querier.CreateMovementRelation(ctx, params)
	if err != nil {
		return domain.MovementRelation{}, translateError(err)
	}
	return mapDBMovementRelationToDomain(dbRel), nil
}

func (r *RelationRepository) CreateEventRelation(ctx context.Context, rel domain.EventRelation) (domain.EventRelation, error) {
	if r.db.Querier == nil {
		return domain.EventRelation{}, ErrUnavailable
	}
	params := db.CreateEventRelationParams{
		ID:           rel.ID,
		EventIDA:     rel.EventIDA,
		EventIDB:     rel.EventIDB,
		RelationType: rel.RelationType,
		Status:       rel.Status,
		Confidence:   toInterfaceFloat64(rel.Confidence),
		Reason:       toInterface(rel.Reason),
		CreatedAt:    rel.CreatedAt,
		UpdatedAt:    rel.UpdatedAt,
	}
	dbRel, err := r.db.Querier.CreateEventRelation(ctx, params)
	if err != nil {
		return domain.EventRelation{}, translateError(err)
	}
	return mapDBEventRelationToDomain(dbRel), nil
}

func (r *RelationRepository) ListMovementRelations(ctx context.Context, movementID string, direction string) ([]domain.MovementRelation, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	// Simplified: list by movement_id_a
	params := db.ListRelationsByMovementParams{
		MovementIDA: movementID,
		MovementIDB: movementID,
	}
	dbRels, err := r.db.Querier.ListRelationsByMovement(ctx, params)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.MovementRelation, 0, len(dbRels))
	for _, rel := range dbRels {
		out = append(out, mapDBMovementRelationToDomain(rel))
	}
	return out, nil
}

func (r *RelationRepository) ListEventRelations(ctx context.Context, eventID string) ([]domain.EventRelation, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	params := db.ListRelationsByEventParams{
		EventIDA: eventID,
		EventIDB: eventID,
	}
	dbRels, err := r.db.Querier.ListRelationsByEvent(ctx, params)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.EventRelation, 0, len(dbRels))
	for _, rel := range dbRels {
		out = append(out, mapDBEventRelationToDomain(rel))
	}
	return out, nil
}

func mapDBMovementRelationToDomain(r db.MovementRelation) domain.MovementRelation {
	floatPtr := func(v interface{}) *float64 {
		if v == nil {
			return nil
		}
		f := v.(float64)
		return &f
	}
	strPtr := func(v interface{}) *string {
		if v == nil {
			return nil
		}
		s := v.(string)
		return &s
	}
	return domain.MovementRelation{
		ID:           r.ID,
		MovementIDA:  r.MovementIDA,
		MovementIDB:  r.MovementIDB,
		RelationType: r.RelationType,
		Status:       r.Status,
		Confidence:   floatPtr(r.Confidence),
		Reason:       strPtr(r.Reason),
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func mapDBEventRelationToDomain(r db.EventRelation) domain.EventRelation {
	floatPtr := func(v interface{}) *float64 {
		if v == nil {
			return nil
		}
		f := v.(float64)
		return &f
	}
	strPtr := func(v interface{}) *string {
		if v == nil {
			return nil
		}
		s := v.(string)
		return &s
	}
	return domain.EventRelation{
		ID:           r.ID,
		EventIDA:     r.EventIDA,
		EventIDB:     r.EventIDB,
		RelationType: r.RelationType,
		Status:       r.Status,
		Confidence:   floatPtr(r.Confidence),
		Reason:       strPtr(r.Reason),
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
