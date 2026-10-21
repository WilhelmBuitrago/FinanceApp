package repositories

import (
	"context"
	"database/sql"
	"errors"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
)

type AccountRepository struct {
	db *DB
}

func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, acc domain.Account) (domain.Account, error) {
	if r.db.Querier == nil {
		return domain.Account{}, ErrUnavailable
	}
	params := db.CreateAccountParams{
		ID:              acc.ID,
		UserID:          acc.UserID,
		Name:            acc.Name,
		AccountType:     acc.AccountType,
		ParentAccountID: toNullString(acc.ParentAccountID),
		Currency:        acc.Currency,
		Status:          acc.Status,
		Institution:     toNullString(acc.Institution),
		MetadataJson:    toNullString(acc.MetadataJson),
		CreatedAt:       acc.CreatedAt,
		UpdatedAt:       acc.UpdatedAt,
	}
	dbAcc, err := r.db.Querier.CreateAccount(ctx, params)
	if err != nil {
		return domain.Account{}, translateError(err)
	}
	return mapDBAccountToDomain(dbAcc), nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (domain.Account, error) {
	if r.db.Querier == nil {
		return domain.Account{}, ErrUnavailable
	}
	dbAcc, err := r.db.Querier.GetAccount(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, ErrNotFound
		}
		return domain.Account{}, translateError(err)
	}
	return mapDBAccountToDomain(dbAcc), nil
}

func (r *AccountRepository) ListByUserID(ctx context.Context, userID string) ([]domain.Account, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbAccs, err := r.db.Querier.ListAccounts(ctx, userID)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Account, 0, len(dbAccs))
	for _, a := range dbAccs {
		out = append(out, mapDBAccountToDomain(a))
	}
	return out, nil
}

func (r *AccountRepository) List(ctx context.Context) ([]domain.Account, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbAccs, err := r.db.Querier.ListAccountsAll(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Account, 0, len(dbAccs))
	for _, a := range dbAccs {
		out = append(out, mapDBAccountToDomain(a))
	}
	return out, nil
}

func mapDBAccountToDomain(a db.Account) domain.Account {
	return domain.Account{
		ID:              a.ID,
		UserID:          a.UserID,
		Name:            a.Name,
		AccountType:     a.AccountType,
		ParentAccountID: fromNullString(a.ParentAccountID),
		Currency:        a.Currency,
		Status:          a.Status,
		Institution:     fromNullString(a.Institution),
		MetadataJson:    fromNullString(a.MetadataJson),
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}
