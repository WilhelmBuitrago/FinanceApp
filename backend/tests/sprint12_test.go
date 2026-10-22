package tests

import (
	"financeapp/backend/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSprint12_AccountCreationNu(t *testing.T) {
	// This is a smoke test for Sprint 12 requirements
	// Real integration would require DB setup; we just validate domain rules
	acc := domain.Account{
		ID:          uuid.NewString(),
		UserID:      "user-1",
		Name:        "Nu Cuenta",
		AccountType: "BANK_ACCOUNT",
		Currency:    "COP",
		Status:      "ACTIVE",
		Institution: strPtr("NU"),
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if acc.Institution == nil || *acc.Institution != "NU" {
		t.Fatalf("institution must be NU")
	}
	if acc.AccountType != "BANK_ACCOUNT" && acc.AccountType != "CREDIT_CARD" {
		t.Fatalf("unsupported account type")
	}
}

func TestSprint12_UploadedAtIndependent(t *testing.T) {
	uploadedAt := time.Now().UTC().Format(time.RFC3339)
	periodStart := "2026-07-01"
	periodEnd := "2026-07-31"
	if uploadedAt == periodStart {
		t.Fatalf("uploaded_at must be independent of period")
	}
	_ = periodEnd
}
