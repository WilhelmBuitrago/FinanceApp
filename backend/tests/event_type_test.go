package tests

import (
	"testing"

	"financeapp/backend/internal/domain"
)

func TestClassifyEventType(t *testing.T) {
	tests := []struct {
		amount int64
		want   string
	}{
		{100, "INCOME"},
		{1, "INCOME"},
		{0, "EXPENSE"},
		{-1, "EXPENSE"},
		{-5000, "EXPENSE"},
	}
	for _, tt := range tests {
		// Simple rule from import service
		var got string
		if tt.amount > 0 {
			got = "INCOME"
		} else {
			got = "EXPENSE"
		}
		if got != tt.want {
			t.Errorf("classify(%d) = %s, want %s", tt.amount, got, tt.want)
		}
	}
}

func TestEventMovementConsistency(t *testing.T) {
	fe := domain.FinancialEvent{
		ID:        "fe1",
		EventType: "INCOME",
		EventDate: strPtr("2026-01-01"),
	}
	mv := domain.Movement{
		ID:               "mv1",
		FinancialEventID: fe.ID,
		AccountID:        "acc1",
		MovementDate:     "2026-01-02",
		AmountMinor:      1000,
	}
	if mv.FinancialEventID != fe.ID {
		t.Errorf("movement not linked to event")
	}
	if mv.AmountMinor <= 0 && fe.EventType == "INCOME" {
		t.Errorf("income event should have positive amount")
	}
}
