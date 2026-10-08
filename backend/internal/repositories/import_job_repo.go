package repositories

import (
	"context"
	"database/sql"
	"financeapp/backend/internal/domain/entities"
)

type ImportJobRepo struct {
	db *sql.DB
}

func NewImportJobRepo(db *sql.DB) *ImportJobRepo {
	return &ImportJobRepo{db: db}
}

func (r *ImportJobRepo) Create(ctx context.Context, job entities.ImportJob) (entities.ImportJob, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO import_jobs(id, document_id, status, created_at, processing_id) VALUES(?,?,?,?,?)`,
		job.ID, job.DocumentID, job.Status, job.CreatedAt, job.ProcessingID)
	if err != nil {
		return entities.ImportJob{}, err
	}
	return job, nil
}

func (r *ImportJobRepo) GetByID(ctx context.Context, id string) (entities.ImportJob, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, document_id, status, created_at, started_at, finished_at, error, processing_id FROM import_jobs WHERE id=?`, id)
	var job entities.ImportJob
	var started, finished, errStr sql.NullString
	var processing sql.NullString
	if err := row.Scan(&job.ID, &job.DocumentID, &job.Status, &job.CreatedAt, &started, &finished, &errStr, &processing); err != nil {
		return entities.ImportJob{}, err
	}
	if started.Valid { s := started.String; job.StartedAt = &s }
	if finished.Valid { s := finished.String; job.FinishedAt = &s }
	if errStr.Valid { s := errStr.String; job.Error = &s }
	if processing.Valid { s := processing.String; job.ProcessingID = &s }
	return job, nil
}

func (r *ImportJobRepo) Update(ctx context.Context, job entities.ImportJob) (entities.ImportJob, error) {
	_, err := r.db.ExecContext(ctx, `UPDATE import_jobs SET status=?, started_at=?, finished_at=?, error=?, processing_id=? WHERE id=?`,
		job.Status, job.StartedAt, job.FinishedAt, job.Error, job.ProcessingID, job.ID)
	if err != nil {
		return entities.ImportJob{}, err
	}
	return job, nil
}
