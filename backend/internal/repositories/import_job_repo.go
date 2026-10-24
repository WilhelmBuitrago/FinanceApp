package repositories

import (
	"context"
	"database/sql"
	"errors"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
)

type ImportJobRepo struct {
	db *DB
}

func NewImportJobRepo(db *DB) *ImportJobRepo {
	return &ImportJobRepo{db: db}
}

func (r *ImportJobRepo) Create(ctx context.Context, job domain.ImportJob) (domain.ImportJob, error) {
	if r.db.Querier == nil {
		return domain.ImportJob{}, ErrUnavailable
	}
	dbJob, err := r.db.Querier.CreateImportJob(ctx, db.CreateImportJobParams{
		ID:           job.ID,
		DocumentID:   job.DocumentID,
		Status:       job.Status,
		CreatedAt:    job.CreatedAt,
		ProcessingID: toNullString(job.ProcessingID),
	})
	if err != nil {
		return domain.ImportJob{}, translateError(err)
	}
	return mapDBImportJobToEntities(dbJob), nil
}

func (r *ImportJobRepo) GetByID(ctx context.Context, id string) (domain.ImportJob, error) {
	if r.db.Querier == nil {
		return domain.ImportJob{}, ErrUnavailable
	}
	dbJob, err := r.db.Querier.GetImportJob(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ImportJob{}, ErrNotFound
		}
		return domain.ImportJob{}, translateError(err)
	}
	return mapDBImportJobToEntities(dbJob), nil
}

func (r *ImportJobRepo) Update(ctx context.Context, job domain.ImportJob) (domain.ImportJob, error) {
	if r.db.Querier == nil {
		return domain.ImportJob{}, ErrUnavailable
	}
	dbJob, err := r.db.Querier.UpdateImportJob(ctx, db.UpdateImportJobParams{
		ID:           job.ID,
		Status:       job.Status,
		StartedAt:    toNullString(job.StartedAt),
		FinishedAt:   toNullString(job.FinishedAt),
		Error:        toNullString(job.Error),
		ProcessingID: toNullString(job.ProcessingID),
	})
	if err != nil {
		return domain.ImportJob{}, translateError(err)
	}
	return mapDBImportJobToEntities(dbJob), nil
}

func mapDBImportJobToEntities(j db.ImportJob) domain.ImportJob {
	return domain.ImportJob{
		ID:           j.ID,
		DocumentID:   j.DocumentID,
		Status:       j.Status,
		CreatedAt:    j.CreatedAt,
		StartedAt:    fromNullString(j.StartedAt),
		FinishedAt:   fromNullString(j.FinishedAt),
		Error:        fromNullString(j.Error),
		ProcessingID: fromNullString(j.ProcessingID),
	}
}
