package repositories

import (
	"context"
	"testing"
	"time"

	"financeapp/backend/internal/domain"
)

func TestAccount_Create_GetByID_RoundTrip(t *testing.T) {
	p := NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	acc := domain.Account{
		ID:          "acc-001",
		UserID:      "user-123",
		Name:        "Main Checking",
		AccountType: "BANK_ACCOUNT",
		Currency:    "USD",
		Status:      "ACTIVE",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	created, err := p.AccountRepo.Create(ctx, acc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID != acc.ID || created.Name != acc.Name {
		t.Fatalf("created mismatch")
	}
	got, err := p.AccountRepo.GetByID(ctx, acc.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.ID != acc.ID || got.UserID != acc.UserID || got.Currency != acc.Currency {
		t.Fatalf("round-trip mismatch: got %+v", got)
	}
}

func TestAccount_GetByID_NotFound(t *testing.T) {
	p := NewTestPersistence(t)
	ctx := context.Background()
	_, err := p.AccountRepo.GetByID(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAccount_ListByUserID(t *testing.T) {
	p := NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	accounts := []domain.Account{
		{ID: "acc-101", UserID: "user-1", Name: "A", AccountType: "CASH", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now},
		{ID: "acc-102", UserID: "user-1", Name: "B", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now},
		{ID: "acc-103", UserID: "user-2", Name: "C", AccountType: "CASH", Currency: "EUR", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now},
	}
	for _, a := range accounts {
		if _, err := p.AccountRepo.Create(ctx, a); err != nil {
			t.Fatalf("create %s: %v", a.ID, err)
		}
	}
	list, err := p.AccountRepo.ListByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("ListByUserID failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(list))
	}
}

func TestAccount_Update_Invalid(t *testing.T) {
	p := NewTestPersistence(t)
	ctx := context.Background()
	acc := domain.Account{ID: "acc-upd", UserID: "u", Name: "N", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"}
	_, err := p.AccountRepo.Update(ctx, acc)
	if err == nil {
		t.Fatal("expected error for Update")
	}
	if err != ErrInvalid {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestAccount_List_ReturnsEmpty(t *testing.T) {
	p := NewTestPersistence(t)
	ctx := context.Background()
	list, err := p.AccountRepo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d", len(list))
	}
}
