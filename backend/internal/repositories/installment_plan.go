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
		OriginEventID:         toInterface(plan.OriginEventID),
		TotalAmountMinor:      toInterfaceInt64(plan.TotalAmountMinor),
		TotalInstallments:     plan.TotalInstallments,
		PaidInstallments:      plan.PaidInstallments,
		RemainingInstallments: toInterfaceInt64(plan.RemainingInstallments),
		Status:                plan.Status,
		StartDate:             toInterface(plan.StartDate),
		ExpectedEndDate:       toInterface(plan.ExpectedEndDate),
		MetadataJson:          toInterface(plan.MetadataJson),
		CreatedAt:             plan.CreatedAt,
		UpdatedAt:             plan.UpdatedAt,
	}
	dbPlan, err := r.db.Querier.CreateInstallmentPlan(ctx, params)
	if err != nil {
		return domain.InstallmentPlan{}, translateError(err)
	}
	// Create installments
	for _, ins := range installments {
		iParams := db.CreateInstallmentParams{
			ID:                  ins.ID,
			InstallmentPlanID:   dbPlan.ID,
			InstallmentNumber:   ins.InstallmentNumber,
			TotalInstallments:   ins.TotalInstallments,
			ExpectedAmountMinor: toInterfaceInt64(ins.ExpectedAmountMinor),
			CapitalAmountMinor:  toInterfaceInt64(ins.CapitalAmountMinor),
			InterestAmountMinor: toInterfaceInt64(ins.InterestAmountMinor),
			Status:              ins.Status,
			ExpectedDate:        toInterface(ins.ExpectedDate),
			DocumentedDate:      toInterface(ins.DocumentedDate),
			MovementID:          toInterface(ins.MovementID),
			MetadataJson:        toInterface(ins.MetadataJson),
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

func (r *InstallmentPlanRepository) Update(ctx context.Context, plan domain.InstallmentPlan) (domain.InstallmentPlan, error) {
	return domain.InstallmentPlan{}, ErrInvalid
}

func mapDBInstallmentPlanToDomain(p db.InstallmentPlan) domain.InstallmentPlan {
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
		s := v.(string)
		return &s
	}
	return domain.InstallmentPlan{
		ID:                    p.ID,
		AccountID:             p.AccountID,
		OriginEventID:         strPtr(p.OriginEventID),
		TotalAmountMinor:      int64Ptr(p.TotalAmountMinor),
		TotalInstallments:     p.TotalInstallments,
		PaidInstallments:      p.PaidInstallments,
		RemainingInstallments: int64Ptr(p.RemainingInstallments),
		Status:                p.Status,
		StartDate:             strPtr(p.StartDate),
		ExpectedEndDate:       strPtr(p.ExpectedEndDate),
		MetadataJson:          strPtr(p.MetadataJson),
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
}

func mapDBInstallmentToDomain(i db.Installment) domain.Installment {
	int64Ptr := func(v interface{}) *int64 {
		if v == nil {
			return nil
		}
		iv := v.(int64)
		return &iv
	}
	strPtr := func(v interface{}) *string {
		if v == nil {
			return nil
		}
		s := v.(string)
		return &s
	}
	return domain.Installment{
		ID:                  i.ID,
		InstallmentPlanID:   i.InstallmentPlanID,
		InstallmentNumber:   i.InstallmentNumber,
		TotalInstallments:   i.TotalInstallments,
		ExpectedAmountMinor: int64Ptr(i.ExpectedAmountMinor),
		CapitalAmountMinor:  int64Ptr(i.CapitalAmountMinor),
		InterestAmountMinor: int64Ptr(i.InterestAmountMinor),
		Status:              i.Status,
		ExpectedDate:        strPtr(i.ExpectedDate),
		DocumentedDate:      strPtr(i.DocumentedDate),
		MovementID:          strPtr(i.MovementID),
		MetadataJson:        strPtr(i.MetadataJson),
		CreatedAt:           i.CreatedAt,
		UpdatedAt:           i.UpdatedAt,
	}
}
