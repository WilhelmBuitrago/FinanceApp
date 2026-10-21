package repositories

import (
	"context"
	"database/sql"
	"errors"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain/entities"
)

type ImportJobRepo struct {
	db *DB
}

func NewImportJobRepo(db *DB) *ImportJobRepo {
	return &ImportJobRepo{db: db}
}

func (r *ImportJobRepo) Create(ctx context.Context, job entities.ImportJob) (entities.ImportJob, error) {
	if r.db.Querier == nil {
		return entities.ImportJob{}, ErrUnavailable
	}
	dbJob, err := r.db.Querier.CreateImportJob(ctx, db.CreateImportJobParams{
		ID:           job.ID,
		DocumentID:   job.DocumentID,
		Status:       job.Status,
		CreatedAt:    job.CreatedAt,
		ProcessingID: toNullString(job.ProcessingID),
	})
	if err != nil {
		return entities.ImportJob{}, translateError(err)
	}
	return mapDBImportJobToEntities(dbJob), nil
}

func (r *ImportJobRepo) GetByID(ctx context.Context, id string) (entities.ImportJob, error) {
	if r.db.Querier == nil {
		return entities.ImportJob{}, ErrUnavailable
	}
	dbJob, err := r.db.Querier.GetImportJob(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.ImportJob{}, ErrNotFound
		}
		return entities.ImportJob{}, translateError(err)
	}
	return mapDBImportJobToEntities(dbJob), nil
}

func (r *ImportJobRepo) Update(ctx context.Context, job entities.ImportJob) (entities.ImportJob, error) {
	if r.db.Querier == nil {
		return entities.ImportJob{}, ErrUnavailable
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
		return entities.ImportJob{}, translateError(err)
	}
	return mapDBImportJobToEntities(dbJob), nil
}

func mapDBImportJobToEntities(j db.ImportJob) entities.ImportJob {
	return entities.ImportJob{
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
