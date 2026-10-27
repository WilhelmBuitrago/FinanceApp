package tests

import (
	"context"
	"database/sql"
	"testing"

	"financeapp/backend/internal/app"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

/* The transfer service: one decision with two legs. The invariants that
   matter are atomicity (a failed leg leaves no half-transfer) and the
   counterpart-aware delete (a leg deletes its transfer). */

type transferRepos struct {
	accounts  *repositories.AccountRepository
	movements *repositories.MovementRepository
	events    *repositories.FinancialEventRepository
	relations *repositories.RelationRepository
}

func transferEnv(t *testing.T) (*app.TransferService, transferRepos, *sql.DB) {
	t.Helper()
	sqlDB := openTestSQLite(t, "transfer.db")
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	repoDB := &repositories.DB{Querier: db.New(sqlDB)}
	return app.NewTransferService(sqlDB), transferRepos{
		accounts:  repositories.NewAccountRepository(repoDB),
		movements: repositories.NewMovementRepository(repoDB),
		events:    repositories.NewFinancialEventRepository(repoDB),
		relations: repositories.NewRelationRepository(repoDB),
	}, sqlDB
}

func seedAccount(t *testing.T, repos transferRepos, id string) {
	t.Helper()
	if _, err := repos.accounts.Create(context.Background(), domain.Account{
		ID: id, UserID: "u", Name: id, AccountType: "BANK_ACCOUNT",
		Currency: "COP", Status: "ACTIVE", Institution: strPtr("NU"),
		CreatedAt: "2026-10-26T00:00:00Z", UpdatedAt: "2026-10-26T00:00:00Z",
	}); err != nil {
		t.Fatalf("seed account %s: %v", id, err)
	}
}

func transferCmd(origin, destination string, amountMinor int64) domain.TransferCommand {
	return domain.TransferCommand{
		OriginID:      origin,
		DestinationID: destination,
		AmountMinor:   amountMinor,
		Date:          "2026-10-26",
		Description:   "Alquiler",
	}
}

func TestTransfer_CreatesBothLegsOneEventAndTheRelation(t *testing.T) {
	service, repos, _ := transferEnv(t)
	seedAccount(t, repos, "acc-a")
	seedAccount(t, repos, "acc-b")

	out, err := service.Transfer(context.Background(), transferCmd("acc-a", "acc-b", 250000))
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if out.EventID == "" || out.OriginMovementID == "" || out.DestinationMovementID == "" || out.RelationID == "" {
		t.Fatalf("the outcome must carry every identity: %+v", out)
	}
	if out.OriginMovementID == out.DestinationMovementID {
		t.Fatal("the two legs must be different movements")
	}

	origin, err := repos.movements.GetByID(context.Background(), out.OriginMovementID)
	if err != nil {
		t.Fatalf("origin leg missing: %v", err)
	}
	if origin.AccountID != "acc-a" || origin.AmountMinor != -250000 {
		t.Errorf("origin leg must be the negative amount in the origin account, got %s %d", origin.AccountID, origin.AmountMinor)
	}
	destination, err := repos.movements.GetByID(context.Background(), out.DestinationMovementID)
	if err != nil {
		t.Fatalf("destination leg missing: %v", err)
	}
	if destination.AccountID != "acc-b" || destination.AmountMinor != 250000 {
		t.Errorf("destination leg must be the positive amount in the destination account, got %s %d", destination.AccountID, destination.AmountMinor)
	}
	if origin.FinancialEventID != out.EventID || destination.FinancialEventID != out.EventID {
		t.Error("both legs must share one financial event")
	}

	event, err := repos.events.GetByID(context.Background(), out.EventID)
	if err != nil {
		t.Fatalf("event missing: %v", err)
	}
	if event.EventType != "TRANSFER" {
		t.Errorf("the event type must be TRANSFER, got %s", event.EventType)
	}

	rels, err := repos.relations.ListMovementRelations(context.Background(), out.OriginMovementID)
	if err != nil {
		t.Fatalf("list relations: %v", err)
	}
	if len(rels) != 1 {
		t.Fatalf("expected exactly one relation, got %d", len(rels))
	}
	if rels[0].RelationType != "TRANSFER_COUNTERPART" || rels[0].Status != "CONFIRMED" {
		t.Errorf("the relation must be a confirmed counterpart, got %s/%s", rels[0].RelationType, rels[0].Status)
	}
	if rels[0].MovementIDB != out.DestinationMovementID {
		t.Error("the relation must bind the origin leg to the destination leg")
	}
}

func TestTransfer_RollsBackEverythingWhenALegFails(t *testing.T) {
	service, repos, sqlDB := transferEnv(t)
	seedAccount(t, repos, "acc-a")
	// No destination account: the second leg fails on referential integrity.

	if _, err := service.Transfer(context.Background(), transferCmd("acc-a", "acc-missing", 250000)); err == nil {
		t.Fatal("a transfer with an unknown destination must fail")
	}

	movements, err := repos.movements.List(context.Background(), 500)
	if err != nil {
		t.Fatalf("list movements: %v", err)
	}
	if len(movements) != 0 {
		t.Errorf("no leg may survive a failed transfer, got %d", len(movements))
	}
	var events int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM financial_events").Scan(&events); err != nil {
		t.Fatalf("count events: %v", err)
	}
	if events != 0 {
		t.Errorf("no event may survive a failed transfer, got %d", events)
	}
	var relations int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM movement_relations").Scan(&relations); err != nil {
		t.Fatalf("count relations: %v", err)
	}
	if relations != 0 {
		t.Errorf("no relation may survive a failed transfer, got %d", relations)
	}
}

func TestTransfer_RejectsTheSameAccountOnBothSides(t *testing.T) {
	service, repos, _ := transferEnv(t)
	seedAccount(t, repos, "acc-a")

	if _, err := service.Transfer(context.Background(), transferCmd("acc-a", "acc-a", 250000)); err == nil {
		t.Fatal("a transfer from an account to itself must be rejected")
	}
	movements, _ := repos.movements.List(context.Background(), 500)
	if len(movements) != 0 {
		t.Error("a rejected transfer must not write anything")
	}
}

func TestTransfer_RejectsANonPositiveAmount(t *testing.T) {
	service, repos, _ := transferEnv(t)
	seedAccount(t, repos, "acc-a")
	seedAccount(t, repos, "acc-b")

	for _, amount := range []int64{0, -250000} {
		if _, err := service.Transfer(context.Background(), transferCmd("acc-a", "acc-b", amount)); err == nil {
			t.Errorf("amount %d must be rejected: the sign is derived, not submitted", amount)
		}
	}
}

func TestTransfer_DeletingALegDeletesTheWholeTransfer(t *testing.T) {
	service, repos, _ := transferEnv(t)
	seedAccount(t, repos, "acc-a")
	seedAccount(t, repos, "acc-b")

	out, err := service.Transfer(context.Background(), transferCmd("acc-a", "acc-b", 250000))
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}

	if err := service.DeleteMovement(context.Background(), out.OriginMovementID); err != nil {
		t.Fatalf("delete a leg: %v", err)
	}

	for _, leg := range []string{out.OriginMovementID, out.DestinationMovementID} {
		if _, err := repos.movements.GetByID(context.Background(), leg); err == nil {
			t.Errorf("leg %s must be gone with its transfer", leg)
		}
	}
	rels, err := repos.relations.ListMovementRelations(context.Background(), out.DestinationMovementID)
	if err != nil {
		t.Fatalf("list relations: %v", err)
	}
	if len(rels) != 0 {
		t.Error("the relation must be gone with its transfer")
	}
}

func TestTransfer_DeletingAPlainMovementStillWorks(t *testing.T) {
	service, repos, _ := transferEnv(t)
	seedAccount(t, repos, "acc-a")

	if _, err := repos.events.Create(context.Background(), domain.FinancialEvent{
		ID: "fe-plain", EventType: "EXPENSE", Status: "ACTIVE",
		VerificationStatus: "UNVERIFIED", CreatedAt: "2026-10-26T00:00:00Z", UpdatedAt: "2026-10-26T00:00:00Z",
	}); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	created, err := repos.movements.Create(context.Background(), domain.Movement{
		ID: "mv-plain", FinancialEventID: "fe-plain", AccountID: "acc-a",
		MovementDate: "2026-10-26", AmountMinor: -1000, Status: "ACTIVE",
		CreatedAt: "2026-10-26T00:00:00Z", UpdatedAt: "2026-10-26T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed plain movement: %v", err)
	}

	if err := service.DeleteMovement(context.Background(), created.ID); err != nil {
		t.Fatalf("delete a plain movement: %v", err)
	}
	if _, err := repos.movements.GetByID(context.Background(), created.ID); err == nil {
		t.Error("the plain movement must be gone")
	}
}
