package tests

import (
	"context"
	"testing"

	"financeapp/backend/internal/domain"
)

/* The orphan event: a movement delete removed the movement and left
   its financial event behind — no movements referencing it, one ghost
   row the listings and the event trace serve forever. The transfer
   path deleted both legs and left the same ghost. The rule is now one
   sentence: the event outlives its last movement never. */

func TestDeleteMovement_RemovesTheEventWhenItsLastLegGoes(t *testing.T) {
	service, repos, sqlDB := transferEnv(t)
	defer func() { _ = sqlDB.Close() }()
	seedAccount(t, repos, "acc-oe1")

	if _, err := repos.events.Create(context.Background(), domain.FinancialEvent{
		ID: "fe-oe1", EventType: "EXPENSE", Status: "ACTIVE",
		VerificationStatus: "UNVERIFIED", CreatedAt: "2026-10-26T00:00:00Z", UpdatedAt: "2026-10-26T00:00:00Z",
	}); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	created, err := repos.movements.Create(context.Background(), domain.Movement{
		ID: "mv-oe1", FinancialEventID: "fe-oe1", AccountID: "acc-oe1",
		MovementDate: "2026-10-26", AmountMinor: -1000, Status: "ACTIVE",
		CreatedAt: "2026-10-26T00:00:00Z", UpdatedAt: "2026-10-26T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed movement: %v", err)
	}

	if err := service.DeleteMovement(context.Background(), created.ID); err != nil {
		t.Fatalf("delete the last leg: %v", err)
	}
	if _, err := repos.events.GetByID(context.Background(), "fe-oe1"); err == nil {
		t.Error("the event must go with its last movement")
	}
}

func TestDeleteMovement_KeepsTheEventWhileAnotherMovementLives(t *testing.T) {
	service, repos, sqlDB := transferEnv(t)
	defer func() { _ = sqlDB.Close() }()
	seedAccount(t, repos, "acc-oe2")

	if _, err := repos.events.Create(context.Background(), domain.FinancialEvent{
		ID: "fe-oe2", EventType: "EXPENSE", Status: "ACTIVE",
		VerificationStatus: "UNVERIFIED", CreatedAt: "2026-10-26T00:00:00Z", UpdatedAt: "2026-10-26T00:00:00Z",
	}); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	one, err := repos.movements.Create(context.Background(), domain.Movement{
		ID: "mv-oe2a", FinancialEventID: "fe-oe2", AccountID: "acc-oe2",
		MovementDate: "2026-10-26", AmountMinor: -1000, Status: "ACTIVE",
		CreatedAt: "2026-10-26T00:00:00Z", UpdatedAt: "2026-10-26T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed movement one: %v", err)
	}
	if _, err := repos.movements.Create(context.Background(), domain.Movement{
		ID: "mv-oe2b", FinancialEventID: "fe-oe2", AccountID: "acc-oe2",
		MovementDate: "2026-10-26", AmountMinor: -1000, Status: "ACTIVE",
		CreatedAt: "2026-10-26T00:00:00Z", UpdatedAt: "2026-10-26T00:00:00Z",
	}); err != nil {
		t.Fatalf("seed movement two: %v", err)
	}

	if err := service.DeleteMovement(context.Background(), one.ID); err != nil {
		t.Fatalf("delete one leg: %v", err)
	}
	if _, err := repos.events.GetByID(context.Background(), "fe-oe2"); err != nil {
		t.Error("the event must survive while a movement still references it")
	}
}

func TestTransfer_DeletingALegLeavesNoOrphanEvent(t *testing.T) {
	service, repos, sqlDB := transferEnv(t)
	defer func() { _ = sqlDB.Close() }()
	seedAccount(t, repos, "acc-oe3a")
	seedAccount(t, repos, "acc-oe3b")

	out, err := service.Transfer(context.Background(), transferCmd("acc-oe3a", "acc-oe3b", 250000))
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if err := service.DeleteMovement(context.Background(), out.OriginMovementID); err != nil {
		t.Fatalf("delete a transfer leg: %v", err)
	}
	if _, err := repos.events.GetByID(context.Background(), out.EventID); err == nil {
		t.Error("a deleted transfer must leave no ghost event behind")
	}
}
