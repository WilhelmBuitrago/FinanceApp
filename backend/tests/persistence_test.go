package tests

import (
	"financeapp/backend/internal/repositories"
	"context"
	"testing"
	"time"

	"financeapp/backend/internal/domain"
)

func TestPersistence_Smoke_FullChain(t *testing.T) {
	p := repositories.NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	acc := domain.Account{ID: "smoke-acc", UserID: "u-smoke", Name: "Smoke", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil { t.Fatalf("account: %v", err) }

	fe := domain.FinancialEvent{ID: "smoke-fe", EventType: "EXPENSE", Status: "ACTIVE", VerificationStatus: "VERIFIED", CreatedAt: now, UpdatedAt: now}
	if _, err := p.FinancialEventRepo.Create(ctx, fe); err != nil { t.Fatalf("event: %v", err) }

	mv := domain.Movement{ID: "smoke-mv", FinancialEventID: "smoke-fe", AccountID: "smoke-acc", MovementDate: "2026-09-22", AmountMinor: 9999, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.MovementRepo.Create(ctx, mv, nil); err != nil { t.Fatalf("movement: %v", err) }

	plan := domain.InstallmentPlan{ID: "smoke-plan", AccountID: "smoke-acc", TotalInstallments: 2, PaidInstallments: 0, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	installments := []domain.Installment{
		{ID: "smoke-ins-1", InstallmentNumber: 1, TotalInstallments: 2, Status: "EXPECTED", CreatedAt: now, UpdatedAt: now},
		{ID: "smoke-ins-2", InstallmentNumber: 2, TotalInstallments: 2, Status: "EXPECTED", CreatedAt: now, UpdatedAt: now},
	}
	if _, err := p.InstallmentPlanRepo.Create(ctx, plan, installments); err != nil { t.Fatalf("plan: %v", err) }

	rel := domain.MovementRelation{ID: "smoke-rel", MovementIDA: "smoke-mv", MovementIDB: "smoke-mv", RelationType: "RELATED_TO", Status: "CANDIDATE", CreatedAt: now, UpdatedAt: now}
	// Note: CHECK movement_id_a <> movement_id_b will fail, use different IDs
	mv2 := domain.Movement{ID: "smoke-mv-2", FinancialEventID: "smoke-fe", AccountID: "smoke-acc", MovementDate: "2026-09-23", AmountMinor: 1111, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.MovementRepo.Create(ctx, mv2, nil); err != nil { t.Fatalf("mv2: %v", err) }
	rel.MovementIDB = "smoke-mv-2"
	if _, err := p.RelationRepo.CreateMovementRelation(ctx, rel); err != nil { t.Fatalf("relation: %v", err) }

	// Verify retrieval
	if _, err := p.AccountRepo.GetByID(ctx, "smoke-acc"); err != nil { t.Fatalf("get acc: %v", err) }
	if _, err := p.MovementRepo.GetByID(ctx, "smoke-mv"); err != nil { t.Fatalf("get mv: %v", err) }
	if _, err := p.InstallmentPlanRepo.GetByID(ctx, "smoke-plan"); err != nil { t.Fatalf("get plan: %v", err) }
}
