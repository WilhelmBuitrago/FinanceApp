package repositories

import (
	"context"
	"database/sql"
	"errors"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
)

type DocumentRepository struct {
	db *DB
}

func NewDocumentRepository(db *DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, doc domain.Document) (domain.Document, error) {
	if r.db.Querier == nil {
		return domain.Document{}, ErrUnavailable
	}
	params := db.CreateDocumentParams{
		ID:           doc.ID,
		AccountID:    toInterface(doc.AccountID),
		DocumentType: doc.DocumentType,
		Institution:  toInterface(doc.Institution),
		DocumentName: toInterface(doc.DocumentName),
		PeriodStart:  toInterface(doc.PeriodStart),
		PeriodEnd:    toInterface(doc.PeriodEnd),
		DocumentDate: toInterface(doc.DocumentDate),
		FileHash:     toInterface(doc.FileHash),
		SourceUri:    toInterface(doc.SourceUri),
		Status:       doc.Status,
		MetadataJson: toInterface(doc.MetadataJson),
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
	dbDoc, err := r.db.Querier.CreateDocument(ctx, params)
	if err != nil {
		return domain.Document{}, translateError(err)
	}
	return mapDBDocumentToDomain(dbDoc), nil
}

func (r *DocumentRepository) GetByID(ctx context.Context, id string) (domain.Document, error) {
	if r.db.Querier == nil {
		return domain.Document{}, ErrUnavailable
	}
	dbDoc, err := r.db.Querier.GetDocument(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Document{}, ErrNotFound
		}
		return domain.Document{}, translateError(err)
	}
	return mapDBDocumentToDomain(dbDoc), nil
}

func (r *DocumentRepository) ListByAccount(ctx context.Context, accountID string) ([]domain.Document, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbDocs, err := r.db.Querier.ListDocumentsByAccount(ctx, accountID)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Document, 0, len(dbDocs))
	for _, d := range dbDocs {
		out = append(out, mapDBDocumentToDomain(d))
	}
	return out, nil
}

func (r *DocumentRepository) List(ctx context.Context) ([]domain.Document, error) {
	if r.db.Querier == nil {
		return nil, ErrUnavailable
	}
	dbDocs, err := r.db.Querier.ListDocumentsAll(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	out := make([]domain.Document, 0, len(dbDocs))
	for _, d := range dbDocs {
		out = append(out, mapDBDocumentToDomain(d))
	}
	return out, nil
}

func (r *DocumentRepository) Update(ctx context.Context, doc domain.Document) (domain.Document, error) {
	return domain.Document{}, ErrInvalid
}

func mapDBDocumentToDomain(d db.Document) domain.Document {
	strPtr := func(v interface{}) *string {
		if v == nil {
			return nil
		}
		s := v.(string)
		return &s
	}
	return domain.Document{
		ID:           d.ID,
		AccountID:    strPtr(d.AccountID),
		DocumentType: d.DocumentType,
		Institution:  strPtr(d.Institution),
		DocumentName: strPtr(d.DocumentName),
		PeriodStart:  strPtr(d.PeriodStart),
		PeriodEnd:    strPtr(d.PeriodEnd),
		DocumentDate: strPtr(d.DocumentDate),
		FileHash:     strPtr(d.FileHash),
		SourceUri:    strPtr(d.SourceUri),
		Status:       d.Status,
		MetadataJson: strPtr(d.MetadataJson),
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
