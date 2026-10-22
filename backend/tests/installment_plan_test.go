package tests

import (
	"context"
	"errors"
	"financeapp/backend/internal/repositories"
	"testing"
	"time"

	"financeapp/backend/internal/domain"
)

func TestInstallmentPlan_Create_WithInstallments_RoundTrip(t *testing.T) {
	p := repositories.NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	acc := domain.Account{ID: "acc-ip-1", UserID: "u", Name: "Acc", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("acc: %v", err)
	}

	total := int64(30000)
	plan := domain.InstallmentPlan{
		ID:                "plan-001",
		AccountID:         "acc-ip-1",
		TotalAmountMinor:  &total,
		TotalInstallments: 3,
		PaidInstallments:  0,
		Status:            "ACTIVE",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	installments := []domain.Installment{
		{ID: "ins-001", InstallmentNumber: 1, TotalInstallments: 3, ExpectedAmountMinor: int64Ptr(10000), Status: "EXPECTED", CreatedAt: now, UpdatedAt: now},
		{ID: "ins-002", InstallmentNumber: 2, TotalInstallments: 3, ExpectedAmountMinor: int64Ptr(10000), Status: "EXPECTED", CreatedAt: now, UpdatedAt: now},
		{ID: "ins-003", InstallmentNumber: 3, TotalInstallments: 3, ExpectedAmountMinor: int64Ptr(10000), Status: "EXPECTED", CreatedAt: now, UpdatedAt: now},
	}
	created, err := p.InstallmentPlanRepo.Create(ctx, plan, installments)
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if created.ID != plan.ID {
		t.Fatalf("id mismatch")
	}

	got, err := p.InstallmentPlanRepo.GetByID(ctx, plan.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.TotalInstallments != 3 {
		t.Fatalf("total installments mismatch")
	}

	ins, err := p.InstallmentPlanRepo.ListInstallments(ctx, plan.ID)
	if err != nil {
		t.Fatalf("list installments: %v", err)
	}
	if len(ins) != 3 {
		t.Fatalf("expected 3 installments, got %d", len(ins))
	}
}

// TestInstallmentPlan_DuplicateNumber_Conflict pins the contract: a duplicate
// installment number inside a plan is rejected with repositories.ErrConflict. The unique
// index is the backstop; this test keeps the taxonomy stable regardless of it.
func TestInstallmentPlan_DuplicateNumber_Conflict(t *testing.T) {
	p := repositories.NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	acc := domain.Account{ID: "acc-dup-num", UserID: "u", Name: "Acc", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("acc: %v", err)
	}

	plan := domain.InstallmentPlan{
		ID:                "plan-dup-num",
		AccountID:         "acc-dup-num",
		TotalInstallments: 3,
		PaidInstallments:  0,
		Status:            "ACTIVE",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	installments := []domain.Installment{
		{ID: "dup-1", InstallmentNumber: 1, TotalInstallments: 3, Status: "EXPECTED", CreatedAt: now, UpdatedAt: now},
		{ID: "dup-2", InstallmentNumber: 2, TotalInstallments: 3, Status: "EXPECTED", CreatedAt: now, UpdatedAt: now},
		{ID: "dup-3", InstallmentNumber: 2, TotalInstallments: 3, Status: "EXPECTED", CreatedAt: now, UpdatedAt: now},
	}
	if _, err := p.InstallmentPlanRepo.Create(ctx, plan, installments); !errors.Is(err, repositories.ErrConflict) {
		t.Fatalf("expected repositories.ErrConflict for duplicate installment number, got %v", err)
	}
}
