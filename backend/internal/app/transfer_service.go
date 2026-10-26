package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

// TransferService owns transfers between own accounts: one financial
// event of type TRANSFER with two movements — the amount leaves the
// origin and lands in the destination — bound by a TRANSFER_COUNTERPART
// relation, all in a single transaction. The RelationRepository this
// uses was scaffolding; this is the feature it was designed for.
//
// A transfer never reaches the review queue: the queue models a
// correction to one movement, not a decision with two legs.
type TransferService struct {
	db *sql.DB
}

func NewTransferService(db *sql.DB) *TransferService {
	return &TransferService{db: db}
}

// Transfer writes both legs, their event and their relation atomically.
func (s *TransferService) Transfer(ctx context.Context, cmd domain.TransferCommand) (out domain.TransferOutcome, err error) {
	if cmd.OriginID == "" || cmd.DestinationID == "" {
		return domain.TransferOutcome{}, fmt.Errorf("transfer needs an origin and a destination account")
	}
	if cmd.OriginID == cmd.DestinationID {
		return domain.TransferOutcome{}, domain.ErrTransferToSelf
	}
	if cmd.AmountMinor <= 0 {
		return domain.TransferOutcome{}, domain.ErrTransferAmount
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.TransferOutcome{}, fmt.Errorf("begin tx: %w", err)
	}
	// Named return: any failure below arms the rollback, so a failed leg
	// never leaves a half-transfer behind.
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	repoDB := &repositories.DB{Querier: db.New(tx)}
	eventRepo := repositories.NewFinancialEventRepository(repoDB)
	movementRepo := repositories.NewMovementRepository(repoDB)
	relationRepo := repositories.NewRelationRepository(repoDB)

	now := time.Now().UTC().Format(time.RFC3339)
	nonce := time.Now().UnixNano()
	eventID := fmt.Sprintf("fe_transfer_%d", nonce)
	originID := fmt.Sprintf("mv_transfer_out_%d", nonce)
	destinationID := fmt.Sprintf("mv_transfer_in_%d", nonce)

	description := cmd.Description
	event := domain.FinancialEvent{
		ID:                 eventID,
		EventType:          "TRANSFER",
		EventDate:          &cmd.Date,
		Status:             "ACTIVE",
		VerificationStatus: "UNVERIFIED",
		Description:        &description,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if _, err := eventRepo.Create(ctx, event); err != nil {
		return domain.TransferOutcome{}, err
	}

	leg := func(id, accountID string, amountMinor int64) domain.Movement {
		return domain.Movement{
			ID:               id,
			FinancialEventID: eventID,
			AccountID:        accountID,
			MovementDate:     cmd.Date,
			AmountMinor:      amountMinor,
			Description:      &description,
			Status:           "ACTIVE",
			CreatedAt:        now,
			UpdatedAt:        now,
		}
	}
	if _, err := movementRepo.Create(ctx, leg(originID, cmd.OriginID, -cmd.AmountMinor)); err != nil {
		return domain.TransferOutcome{}, err
	}
	if _, err := movementRepo.Create(ctx, leg(destinationID, cmd.DestinationID, cmd.AmountMinor)); err != nil {
		return domain.TransferOutcome{}, err
	}

	relation := domain.MovementRelation{
		ID:           fmt.Sprintf("rel_transfer_%d", nonce),
		MovementIDA:  originID,
		MovementIDB:  destinationID,
		RelationType: "TRANSFER_COUNTERPART",
		Status:       "CONFIRMED",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := relationRepo.CreateMovementRelation(ctx, relation); err != nil {
		return domain.TransferOutcome{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.TransferOutcome{}, fmt.Errorf("commit tx: %w", err)
	}
	return domain.TransferOutcome{
		EventID:               eventID,
		OriginMovementID:      originID,
		DestinationMovementID: destinationID,
		RelationID:            relation.ID,
	}, nil
}

// DeleteMovement removes a movement, transfer-aware: deleting either leg
// deletes both legs and the relation that binds them, because the
// transfer is one decision. A movement with no counterpart follows the
// ordinary path.
func (s *TransferService) DeleteMovement(ctx context.Context, movementID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	repoDB := &repositories.DB{Querier: db.New(tx)}
	relationRepo := repositories.NewRelationRepository(repoDB)
	movementRepo := repositories.NewMovementRepository(repoDB)

	relations, err := relationRepo.ListMovementRelations(ctx, movementID)
	if err != nil {
		return err
	}
	for _, rel := range relations {
		if rel.RelationType != "TRANSFER_COUNTERPART" {
			continue
		}
		counterpart := rel.MovementIDB
		if counterpart == movementID {
			counterpart = rel.MovementIDA
		}
		if err := relationRepo.DeleteMovementRelation(ctx, rel.ID); err != nil {
			return err
		}
		if err := movementRepo.Delete(ctx, counterpart); err != nil {
			return err
		}
	}
	if err := movementRepo.Delete(ctx, movementID); err != nil {
		return err
	}
	return tx.Commit()
}
