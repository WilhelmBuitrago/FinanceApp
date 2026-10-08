package repositories

import (
	"context"
	"financeapp/backend/internal/domain/entities"
)

// ImportJobRepository defines persistence operations for import jobs.
type ImportJobRepository interface {
	Create(ctx context.Context, job entities.ImportJob) (entities.ImportJob, error)
	GetByID(ctx context.Context, id string) (entities.ImportJob, error)
	GetByDocumentID(ctx context.Context, documentID string) (entities.ImportJob, error)
	Update(ctx context.Context, job entities.ImportJob) (entities.ImportJob, error)
	ListByStatus(ctx context.Context, status string) ([]entities.ImportJob, error)
}

// TODO: Implement concrete repository with SQLite/sqlc.
// TODO: Ensure transaction safety for status transitions.
