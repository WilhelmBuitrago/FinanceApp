package tests

import (
	"context"
	"testing"
	"time"

	"financeapp/backend/internal/app"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

// TestMovementService_RollbackOnComponentFailure pins that an invalid
// component type aborts the whole write: no orphan movement, no orphan
// financial event. This is the guarantee the credit card breakdown depends on.
func TestMovementService_RollbackOnComponentFailure(t *testing.T) {
	p := newTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	acc := domain.Account{ID: "acc-tx", UserID: "u", Name: "A", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("account: %v", err)
	}

	fe := domain.FinancialEvent{ID: "fe-tx", EventType: "EXPENSE", Status: "ACTIVE", VerificationStatus: "VERIFIED", CreatedAt: now, UpdatedAt: now}
	mv := domain.Movement{ID: "mv-tx", FinancialEventID: "fe-tx", AccountID: "acc-tx", MovementDate: "2026-09-01", AmountMinor: 5000, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	badComp := []domain.MovementComponent{
		{ID: "comp-bad", ComponentType: "INVALID_TYPE", AmountMinor: 5000, CreatedAt: now, UpdatedAt: now},
	}

	service := app.NewMovementService(p.DB, nil)
	if _, err := service.CreateWithEvent(ctx, fe, mv, badComp); err == nil {
		t.Fatal("expected error for invalid component type")
	}

	if _, err := p.MovementRepo.GetByID(ctx, mv.ID); err != repositories.ErrNotFound {
		t.Fatalf("movement must not survive a failed transaction, got %v", err)
	}
	if _, err := p.FinancialEventRepo.GetByID(ctx, fe.ID); err != repositories.ErrNotFound {
		t.Fatalf("financial event must not survive a failed transaction, got %v", err)
	}
}

// TestMovementService_CreatesEventMovementAndComponents pins the happy path:
// event, movement and components commit together and the movement gets its
// components linked.
func TestMovementService_CreatesEventMovementAndComponents(t *testing.T) {
	p := newTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	acc := domain.Account{ID: "acc-ok", UserID: "u", Name: "A", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("account: %v", err)
	}

	fe := domain.FinancialEvent{ID: "fe-ok", EventType: "EXPENSE", Status: "ACTIVE", VerificationStatus: "UNVERIFIED", CreatedAt: now, UpdatedAt: now}
	mv := domain.Movement{ID: "mv-ok", FinancialEventID: "fe-ok", AccountID: "acc-ok", MovementDate: "2026-09-02", AmountMinor: 15000, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	comps := []domain.MovementComponent{
		{ID: "ok-c1", ComponentType: "CAPITAL", AmountMinor: 10000, CreatedAt: now, UpdatedAt: now},
		{ID: "ok-c2", ComponentType: "INSURANCE", AmountMinor: 5000, CreatedAt: now, UpdatedAt: now},
	}

	service := app.NewMovementService(p.DB, nil)
	created, err := service.CreateWithEvent(ctx, fe, mv, comps)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID != mv.ID {
		t.Fatalf("movement id mismatch: %s", created.ID)
	}

	got, err := p.MovementRepo.ListComponents(ctx, mv.ID)
	if err != nil {
		t.Fatalf("list components: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 components, got %d", len(got))
	}
	for _, c := range got {
		if c.MovementID != mv.ID {
			t.Fatalf("component %s not linked to movement (has %s)", c.ID, c.MovementID)
		}
	}
}

func TestTransaction_RollbackOnInstallmentFailure(t *testing.T) {
	p := newTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	acc := domain.Account{ID: "acc-ip-tx", UserID: "u", Name: "A", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("%v", err)
	}

	plan := domain.InstallmentPlan{ID: "plan-tx", AccountID: "acc-ip-tx", TotalInstallments: 2, PaidInstallments: 0, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	// Duplicate installment_number triggers UNIQUE constraint
	badInst := []domain.Installment{
		{ID: "ins-tx-1", InstallmentNumber: 1, TotalInstallments: 2, Status: "EXPECTED", CreatedAt: now, UpdatedAt: now},
		{ID: "ins-tx-2", InstallmentNumber: 1, TotalInstallments: 2, Status: "EXPECTED", CreatedAt: now, UpdatedAt: now},
	}
	_, err := p.InstallmentPlanRepo.Create(ctx, plan, badInst)
	if err == nil {
		t.Fatal("expected error for duplicate installment number")
	}
}
