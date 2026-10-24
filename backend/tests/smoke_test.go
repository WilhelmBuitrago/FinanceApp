package tests

import (
	"context"
	"testing"

	"financeapp/backend/internal/domain"
)

// TestSmokeAccountCreateGet is the end-to-end smoke check: with the real
// schema applied, an account survives a create/get round trip.
func TestSmokeAccountCreateGet(t *testing.T) {
	p := newTestPersistence(t)
	ctx := context.Background()

	acc := domain.Account{
		ID:          "acc-1",
		UserID:      "user-1",
		Name:        "Test Account",
		AccountType: "BANK_ACCOUNT",
		Currency:    "USD",
		Status:      "ACTIVE",
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-01T00:00:00Z",
	}

	created, err := p.AccountRepo.Create(ctx, acc)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	if created.ID != acc.ID {
		t.Fatalf("created ID mismatch: %s vs %s", created.ID, acc.ID)
	}

	fetched, err := p.AccountRepo.GetByID(ctx, acc.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if fetched.Name != acc.Name {
		t.Fatalf("fetched name mismatch: %s vs %s", fetched.Name, acc.Name)
	}
}
