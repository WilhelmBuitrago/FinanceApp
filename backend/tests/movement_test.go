package tests

import (
	"context"
	"errors"
	"financeapp/backend/internal/repositories"
	"testing"
	"time"

	"financeapp/backend/internal/domain"
)

func TestMovement_Create_WithComponents_RoundTrip(t *testing.T) {
	p := repositories.NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	// prerequisites
	acc := domain.Account{ID: "acc-mv-1", UserID: "u1", Name: "Acc", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("account create: %v", err)
	}
	fe := domain.FinancialEvent{ID: "fe-mv-1", EventType: "EXPENSE", Status: "ACTIVE", VerificationStatus: "VERIFIED", CreatedAt: now, UpdatedAt: now}
	if _, err := p.FinancialEventRepo.Create(ctx, fe); err != nil {
		t.Fatalf("event create: %v", err)
	}

	mv := domain.Movement{
		ID:               "mv-001",
		FinancialEventID: "fe-mv-1",
		AccountID:        "acc-mv-1",
		MovementDate:     "2026-09-01",
		AmountMinor:      12345,
		Status:           "ACTIVE",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	comps := []domain.MovementComponent{
		{ID: "comp-001", MovementID: "mv-001", ComponentType: "CAPITAL", AmountMinor: 10000, CreatedAt: now, UpdatedAt: now},
		{ID: "comp-002", MovementID: "mv-001", ComponentType: "INTEREST", AmountMinor: 2345, CreatedAt: now, UpdatedAt: now},
	}
	created, err := p.MovementRepo.Create(ctx, mv)
	if err != nil {
		t.Fatalf("movement create: %v", err)
	}
	for _, c := range comps {
		c.MovementID = created.ID
		if _, err := p.MovementRepo.CreateComponent(ctx, c); err != nil {
			t.Fatalf("component %s rejected: %v", c.ComponentType, err)
		}
	}
	if created.ID != mv.ID {
		t.Fatalf("id mismatch")
	}
	got, err := p.MovementRepo.GetByID(ctx, mv.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.AmountMinor != mv.AmountMinor || got.MovementDate != mv.MovementDate {
		t.Fatalf("round-trip mismatch")
	}
	compsGot, err := p.MovementRepo.ListComponents(ctx, mv.ID)
	if err != nil {
		t.Fatalf("list components: %v", err)
	}
	if len(compsGot) != 2 {
		t.Fatalf("expected 2 components, got %d", len(compsGot))
	}
}

func TestMovement_ForeignKeyViolation(t *testing.T) {
	p := repositories.NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	fe := domain.FinancialEvent{ID: "fe-fk-1", EventType: "INCOME", Status: "ACTIVE", VerificationStatus: "VERIFIED", CreatedAt: now, UpdatedAt: now}
	if _, err := p.FinancialEventRepo.Create(ctx, fe); err != nil {
		t.Fatalf("event: %v", err)
	}

	mv := domain.Movement{
		ID:               "mv-fk-1",
		FinancialEventID: "fe-fk-1",
		AccountID:        "nonexistent-acc",
		MovementDate:     "2026-09-01",
		AmountMinor:      1000,
		Status:           "ACTIVE",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	_, err := p.MovementRepo.Create(ctx, mv)
	if err == nil {
		t.Fatal("expected foreign key error")
	}
	// translateError maps constraint violation
	if err != repositories.ErrConstraint && err != repositories.ErrUnavailable {
		t.Logf("got error %v", err)
	}
}

func TestMovement_DuplicateDocumentaryValues_Independence(t *testing.T) {
	p := repositories.NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	acc := domain.Account{ID: "acc-dup", UserID: "u", Name: "Acc", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("acc: %v", err)
	}
	fe1 := domain.FinancialEvent{ID: "fe-dup-1", EventType: "EXPENSE", Status: "ACTIVE", VerificationStatus: "VERIFIED", CreatedAt: now, UpdatedAt: now}
	fe2 := domain.FinancialEvent{ID: "fe-dup-2", EventType: "EXPENSE", Status: "ACTIVE", VerificationStatus: "VERIFIED", CreatedAt: now, UpdatedAt: now}
	if _, err := p.FinancialEventRepo.Create(ctx, fe1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err := p.FinancialEventRepo.Create(ctx, fe2); err != nil {
		t.Fatalf("%v", err)
	}

	mv1 := domain.Movement{ID: "mv-dup-1", FinancialEventID: "fe-dup-1", AccountID: "acc-dup", MovementDate: "2026-09-01", AmountMinor: 5000, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	mv2 := domain.Movement{ID: "mv-dup-2", FinancialEventID: "fe-dup-2", AccountID: "acc-dup", MovementDate: "2026-09-01", AmountMinor: 5000, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.MovementRepo.Create(ctx, mv1); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err := p.MovementRepo.Create(ctx, mv2); err != nil {
		t.Fatalf("%v", err)
	}

	g1, _ := p.MovementRepo.GetByID(ctx, "mv-dup-1")
	g2, _ := p.MovementRepo.GetByID(ctx, "mv-dup-2")
	if g1.ID == g2.ID {
		t.Fatal("ids collided")
	}
	if g1.FinancialEventID != "fe-dup-1" || g2.FinancialEventID != "fe-dup-2" {
		t.Fatal("independence broken")
	}
}

// TestMovement_ComponentTypes_NuCatalog pins the credit-card component catalog:
// capital, interest, fee (cuota de manejo), seguro and mora. Nu cards are
// separate instruments bound to an account, so the breakdown must cover the
// five types this issuer produces.
func TestMovement_ComponentTypes_NuCatalog(t *testing.T) {
	p := repositories.NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	acc := domain.Account{ID: "acc-nu-types", UserID: "u", Name: "Nu", AccountType: "CREDIT_CARD", Currency: "COP", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("acc: %v", err)
	}
	fe := domain.FinancialEvent{ID: "fe-nu-types", EventType: "EXPENSE", Status: "ACTIVE", VerificationStatus: "VERIFIED", CreatedAt: now, UpdatedAt: now}
	if _, err := p.FinancialEventRepo.Create(ctx, fe); err != nil {
		t.Fatalf("event: %v", err)
	}

	mv := domain.Movement{ID: "mv-nu-types", FinancialEventID: "fe-nu-types", AccountID: "acc-nu-types", MovementDate: "2026-10-01", AmountMinor: 250000, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	comps := []domain.MovementComponent{
		{ID: "nu-c1", MovementID: "mv-nu-types", ComponentType: "CAPITAL", AmountMinor: 200000, CreatedAt: now, UpdatedAt: now},
		{ID: "nu-c2", MovementID: "mv-nu-types", ComponentType: "INTEREST", AmountMinor: 30000, CreatedAt: now, UpdatedAt: now},
		{ID: "nu-c3", MovementID: "mv-nu-types", ComponentType: "FEE", AmountMinor: 10000, CreatedAt: now, UpdatedAt: now},
		{ID: "nu-c4", MovementID: "mv-nu-types", ComponentType: "INSURANCE", AmountMinor: 8000, CreatedAt: now, UpdatedAt: now},
		{ID: "nu-c5", MovementID: "mv-nu-types", ComponentType: "LATE_FEE", AmountMinor: 2000, CreatedAt: now, UpdatedAt: now},
	}
	created, err := p.MovementRepo.Create(ctx, mv)
	if err != nil {
		t.Fatalf("Nu component catalog rejected: %v", err)
	}
	for _, c := range comps {
		c.MovementID = created.ID
		if _, err := p.MovementRepo.CreateComponent(ctx, c); err != nil {
			t.Fatalf("Nu component catalog rejected: %v", err)
		}
	}
	got, err := p.MovementRepo.ListComponents(ctx, "mv-nu-types")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("expected 5 components, got %d", len(got))
	}

	// Legacy Spanish values are rejected after the normalization to English.
	legacy := domain.Movement{ID: "mv-legacy-type", FinancialEventID: "fe-nu-types", AccountID: "acc-nu-types", MovementDate: "2026-10-02", AmountMinor: 5000, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	legacyComps := []domain.MovementComponent{
		{ID: "legacy-c1", MovementID: "mv-legacy-type", ComponentType: "SEGURO", AmountMinor: 5000, CreatedAt: now, UpdatedAt: now},
	}
	if _, err := p.MovementRepo.Create(ctx, legacy); err != nil {
		t.Fatalf("legacy movement setup: %v", err)
	}
	if _, err := p.MovementRepo.CreateComponent(ctx, legacyComps[0]); !errors.Is(err, repositories.ErrConstraint) {
		t.Fatalf("expected legacy Spanish component type to be rejected with repositories.ErrConstraint, got %v", err)
	}
}
