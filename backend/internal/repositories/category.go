package repositories

import (
	"context"
	"database/sql"
	"errors"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
)

type CategoryRepository struct {
	db *DB
}

func NewCategoryRepository(db *DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, cat domain.Category) (domain.Category, error) {
	if r.db.Querier == nil {
		return domain.Category{}, ErrUnavailable
	}
	params := db.CreateCategoryParams{
		ID:           cat.ID,
		Name:         cat.Name,
		ParentID:     toNullString(cat.ParentID),
		CategoryType: toNullString(cat.CategoryType),
		Status:       cat.Status,
		CreatedAt:    cat.CreatedAt,
		UpdatedAt:    cat.UpdatedAt,
	}
	dbCat, err := r.db.Querier.CreateCategory(ctx, params)
	if err != nil {
		return domain.Category{}, translateError(err)
	}
	return mapDBCategoryToDomain(dbCat), nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id string) (domain.Category, error) {
	if r.db.Querier == nil {
		return domain.Category{}, ErrUnavailable
	}
	dbCat, err := r.db.Querier.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Category{}, ErrNotFound
		}
		return domain.Category{}, translateError(err)
	}
	return mapDBCategoryToDomain(dbCat), nil
}

func (r *CategoryRepository) List(ctx context.Context, status string) ([]domain.Category, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbCats, err := r.db.Querier.ListCategories(ctx, status)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Category, 0, len(dbCats))
	for _, c := range dbCats {
		out = append(out, mapDBCategoryToDomain(c))
	}
	return out, nil
}

func (r *CategoryRepository) Update(ctx context.Context, cat domain.Category) (domain.Category, error) {
	return domain.Category{}, ErrInvalid
}

func mapDBCategoryToDomain(c db.Category) domain.Category {
	return domain.Category{
		ID:           c.ID,
		Name:         c.Name,
		ParentID:     fromNullString(c.ParentID),
		CategoryType: fromNullString(c.CategoryType),
		Status:       c.Status,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
