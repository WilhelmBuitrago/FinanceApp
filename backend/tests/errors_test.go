package tests

import (
	"context"
	"financeapp/backend/internal/repositories"
	"testing"
	"time"

	"financeapp/backend/internal/domain"
)

func TestErrors_Taxonomy_NotFound(t *testing.T) {
	p := repositories.NewTestPersistence(t)
	ctx := context.Background()
	_, err := p.AccountRepo.GetByID(ctx, "does-not-exist")
	if err != repositories.ErrNotFound {
		t.Fatalf("expected repositories.ErrNotFound, got %v", err)
	}
}

func TestErrors_Taxonomy_Constraint(t *testing.T) {
	p := repositories.NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	acc := domain.Account{ID: "acc-const", UserID: "u", Name: "A", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("first create: %v", err)
	}
	// duplicate primary key should yield constraint/conflict
	_, err := p.AccountRepo.Create(ctx, acc)
	if err == nil {
		t.Fatal("expected error on duplicate")
	}
	if err != repositories.ErrConflict && err != repositories.ErrConstraint {
		t.Fatalf("expected repositories.ErrConflict or repositories.ErrConstraint, got %v", err)
	}
}
