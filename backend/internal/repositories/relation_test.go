package repositories

import (
	"context"
	"testing"
	"time"

	"financeapp/backend/internal/domain"
)

func TestRelation_MovementRelation_CreateList(t *testing.T) {
	p := NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	acc := domain.Account{ID: "acc-rel", UserID: "u", Name: "A", AccountType: "BANK_ACCOUNT", Currency: "USD", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.AccountRepo.Create(ctx, acc); err != nil { t.Fatalf("%v", err) }
	fe := domain.FinancialEvent{ID: "fe-rel", EventType: "EXPENSE", Status: "ACTIVE", VerificationStatus: "VERIFIED", CreatedAt: now, UpdatedAt: now}
	if _, err := p.FinancialEventRepo.Create(ctx, fe); err != nil { t.Fatalf("%v", err) }

	mv1 := domain.Movement{ID: "mv-r1", FinancialEventID: "fe-rel", AccountID: "acc-rel", MovementDate: "2026-01-01", AmountMinor: 1000, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	mv2 := domain.Movement{ID: "mv-r2", FinancialEventID: "fe-rel", AccountID: "acc-rel", MovementDate: "2026-01-02", AmountMinor: 1000, Status: "ACTIVE", CreatedAt: now, UpdatedAt: now}
	if _, err := p.MovementRepo.Create(ctx, mv1, nil); err != nil { t.Fatalf("%v", err) }
	if _, err := p.MovementRepo.Create(ctx, mv2, nil); err != nil { t.Fatalf("%v", err) }

	rel := domain.MovementRelation{
		ID: "rel-mv-1", MovementIDA: "mv-r1", MovementIDB: "mv-r2", RelationType: "TRANSFER_COUNTERPART", Status: "CONFIRMED", CreatedAt: now, UpdatedAt: now,
	}
	created, err := p.RelationRepo.CreateMovementRelation(ctx, rel)
	if err != nil { t.Fatalf("create relation: %v", err) }
	if created.ID != rel.ID { t.Fatalf("id mismatch") }

	list, err := p.RelationRepo.ListMovementRelations(ctx, "mv-r1", "")
	if err != nil { t.Fatalf("list: %v", err) }
	if len(list) == 0 { t.Fatalf("expected at least one relation") }
}

func TestRelation_EventRelation_CreateList_States(t *testing.T) {
	p := NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	fe1 := domain.FinancialEvent{ID: "fe-e1", EventType: "INCOME", Status: "ACTIVE", VerificationStatus: "VERIFIED", CreatedAt: now, UpdatedAt: now}
	fe2 := domain.FinancialEvent{ID: "fe-e2", EventType: "EXPENSE", Status: "ACTIVE", VerificationStatus: "VERIFIED", CreatedAt: now, UpdatedAt: now}
	if _, err := p.FinancialEventRepo.Create(ctx, fe1); err != nil { t.Fatalf("%v", err) }
	if _, err := p.FinancialEventRepo.Create(ctx, fe2); err != nil { t.Fatalf("%v", err) }

	rel := domain.EventRelation{
		ID: "rel-e1", EventIDA: "fe-e1", EventIDB: "fe-e2", RelationType: "RELATED_TO", Status: "CANDIDATE", CreatedAt: now, UpdatedAt: now,
	}
	created, err := p.RelationRepo.CreateEventRelation(ctx, rel)
	if err != nil { t.Fatalf("create: %v", err) }
	if created.Status != "CANDIDATE" { t.Fatalf("status mismatch") }

	list, err := p.RelationRepo.ListEventRelations(ctx, "fe-e1")
	if err != nil { t.Fatalf("list: %v", err) }
	if len(list) == 0 { t.Fatalf("expected relation") }
}
