package repositories

import (
	"context"
	"database/sql"
	"errors"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
)

type InstallmentPlanRepository struct {
	db *DB
}

// InstallmentPlanRepository is SCAFFOLDING for the planned credit-card
// feature (purchases paid in installments, with their capital/interest
// breakdown). It is not wired to any handler or service yet: no route
// reads or writes installment plans. Its tests keep the schema and the
// generated queries honest until the feature lands and connects here.
// When that happens, this comment goes away with the wiring.

func NewInstallmentPlanRepository(db *DB) *InstallmentPlanRepository {
	return &InstallmentPlanRepository{db: db}
}

func (r *InstallmentPlanRepository) Create(ctx context.Context, plan domain.InstallmentPlan, installments []domain.Installment) (domain.InstallmentPlan, error) {
	if r.db.Querier == nil {
		return domain.InstallmentPlan{}, ErrUnavailable
	}
	params := db.CreateInstallmentPlanParams{
		ID:                    plan.ID,
		AccountID:             plan.AccountID,
		OriginEventID:         toNullString(plan.OriginEventID),
		TotalAmountMinor:      toNullInt64(plan.TotalAmountMinor),
		TotalInstallments:     plan.TotalInstallments,
		PaidInstallments:      plan.PaidInstallments,
		RemainingInstallments: toNullInt64(plan.RemainingInstallments),
		Status:                plan.Status,
		StartDate:             toNullString(plan.StartDate),
		ExpectedEndDate:       toNullString(plan.ExpectedEndDate),
		MetadataJson:          toNullString(plan.MetadataJson),
		CreatedAt:             plan.CreatedAt,
		UpdatedAt:             plan.UpdatedAt,
	}
	dbPlan, err := r.db.Querier.CreateInstallmentPlan(ctx, params)
	if err != nil {
		return domain.InstallmentPlan{}, translateError(err)
	}
	// One installment number per plan: reject duplicates up front so the
	// caller gets a clear ErrConflict instead of a raw UNIQUE failure, and
	// the request never reaches the database with an invalid set. The unique
	// index remains the backstop.
	existing, err := r.db.Querier.ListInstallmentsByPlan(ctx, dbPlan.ID)
	if err != nil {
		return domain.InstallmentPlan{}, translateError(err)
	}
	seenNumbers := make(map[int64]bool, len(existing)+len(installments))
	for _, e := range existing {
		seenNumbers[e.InstallmentNumber] = true
	}
	// Create installments
	for _, ins := range installments {
		if seenNumbers[ins.InstallmentNumber] {
			return domain.InstallmentPlan{}, ErrConflict
		}
		seenNumbers[ins.InstallmentNumber] = true
		iParams := db.CreateInstallmentParams{
			ID:                  ins.ID,
			InstallmentPlanID:   dbPlan.ID,
			InstallmentNumber:   ins.InstallmentNumber,
			TotalInstallments:   ins.TotalInstallments,
			ExpectedAmountMinor: toNullInt64(ins.ExpectedAmountMinor),
			CapitalAmountMinor:  toNullInt64(ins.CapitalAmountMinor),
			InterestAmountMinor: toNullInt64(ins.InterestAmountMinor),
			Status:              ins.Status,
			ExpectedDate:        toNullString(ins.ExpectedDate),
			DocumentedDate:      toNullString(ins.DocumentedDate),
			MovementID:          toNullString(ins.MovementID),
			MetadataJson:        toNullString(ins.MetadataJson),
			CreatedAt:           ins.CreatedAt,
			UpdatedAt:           ins.UpdatedAt,
		}
		_, err := r.db.Querier.CreateInstallment(ctx, iParams)
		if err != nil {
			return domain.InstallmentPlan{}, translateError(err)
		}
	}
	return mapDBInstallmentPlanToDomain(dbPlan), nil
}

func (r *InstallmentPlanRepository) GetByID(ctx context.Context, id string) (domain.InstallmentPlan, error) {
	if r.db.Querier == nil {
		return domain.InstallmentPlan{}, ErrUnavailable
	}
	dbPlan, err := r.db.Querier.GetInstallmentPlan(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.InstallmentPlan{}, ErrNotFound
		}
		return domain.InstallmentPlan{}, translateError(err)
	}
	return mapDBInstallmentPlanToDomain(dbPlan), nil
}

func (r *InstallmentPlanRepository) ListByAccount(ctx context.Context, accountID string) ([]domain.InstallmentPlan, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbPlans, err := r.db.Querier.ListInstallmentPlansByAccount(ctx, accountID)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.InstallmentPlan, 0, len(dbPlans))
	for _, p := range dbPlans {
		out = append(out, mapDBInstallmentPlanToDomain(p))
	}
	return out, nil
}

func (r *InstallmentPlanRepository) ListInstallments(ctx context.Context, planID string) ([]domain.Installment, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbIns, err := r.db.Querier.ListInstallmentsByPlan(ctx, planID)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Installment, 0, len(dbIns))
	for _, i := range dbIns {
		out = append(out, mapDBInstallmentToDomain(i))
	}
	return out, nil
}

func mapDBInstallmentPlanToDomain(p db.InstallmentPlan) domain.InstallmentPlan {
	return domain.InstallmentPlan{
		ID:                    p.ID,
		AccountID:             p.AccountID,
		OriginEventID:         fromNullString(p.OriginEventID),
		TotalAmountMinor:      fromNullInt64(p.TotalAmountMinor),
		TotalInstallments:     p.TotalInstallments,
		PaidInstallments:      p.PaidInstallments,
		RemainingInstallments: fromNullInt64(p.RemainingInstallments),
		Status:                p.Status,
		StartDate:             fromNullString(p.StartDate),
		ExpectedEndDate:       fromNullString(p.ExpectedEndDate),
		MetadataJson:          fromNullString(p.MetadataJson),
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
}

func mapDBInstallmentToDomain(i db.Installment) domain.Installment {
	return domain.Installment{
		ID:                  i.ID,
		InstallmentPlanID:   i.InstallmentPlanID,
		InstallmentNumber:   i.InstallmentNumber,
		TotalInstallments:   i.TotalInstallments,
		ExpectedAmountMinor: fromNullInt64(i.ExpectedAmountMinor),
		CapitalAmountMinor:  fromNullInt64(i.CapitalAmountMinor),
		InterestAmountMinor: fromNullInt64(i.InterestAmountMinor),
		Status:              i.Status,
		ExpectedDate:        fromNullString(i.ExpectedDate),
		DocumentedDate:      fromNullString(i.DocumentedDate),
		MovementID:          fromNullString(i.MovementID),
		MetadataJson:        fromNullString(i.MetadataJson),
		CreatedAt:           i.CreatedAt,
		UpdatedAt:           i.UpdatedAt,
	}
}
