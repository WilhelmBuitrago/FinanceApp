package repositories

import (
	"context"
	"testing"
	"time"

	"financeapp/backend/internal/domain"
)

func TestErrors_Taxonomy_NotFound(t *testing.T) {
	p := NewTestPersistence(t)
	ctx := context.Background()
	_, err := p.AccountRepo.GetByID(ctx, "does-not-exist")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestErrors_Taxonomy_Invalid(t *testing.T) {
	p := NewTestPersistence(t)
	ctx := context.Background()
	acc := domain.Account{ID: "a", UserID: "u", Name: "n", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"}
	_, err := p.AccountRepo.Update(ctx, acc)
	if err != ErrInvalid {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestErrors_Taxonomy_Constraint(t *testing.T) {
	p := NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	acc := domain.Account{ID: "acc-const", UserID: "u", Name: "A", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil { t.Fatalf("first create: %v", err) }
	// duplicate primary key should yield constraint/conflict
	_, err := p.AccountRepo.Create(ctx, acc)
	if err == nil {
		t.Fatal("expected error on duplicate")
	}
	if err != ErrConflict && err != ErrConstraint {
		t.Fatalf("expected ErrConflict or ErrConstraint, got %v", err)
	}
}
