package tests

import (
	"context"
	"financeapp/backend/internal/repositories"
	"testing"
	"time"

	"financeapp/backend/internal/domain"
)

// TestDocument_Update_PersistsStatus pins that a document status change made
// during the import confirm flow actually reaches the database.
func TestDocument_Update_PersistsStatus(t *testing.T) {
	p := repositories.NewTestPersistence(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	doc := domain.Document{
		ID:           "doc-upd",
		DocumentType: "STATEMENT",
		Status:       "UPLOADED",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := p.DocumentRepo.Create(ctx, doc); err != nil {
		t.Fatalf("create: %v", err)
	}

	doc.Status = "PROCESSED"
	doc.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	updated, err := p.DocumentRepo.Update(ctx, doc)
	if err != nil {
		t.Fatalf("update rejected: %v", err)
	}
	if updated.Status != "PROCESSED" {
		t.Fatalf("returned status %q, want PROCESSED", updated.Status)
	}

	got, err := p.DocumentRepo.GetByID(ctx, "doc-upd")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != "PROCESSED" {
		t.Fatalf("status not persisted: %q, want PROCESSED", got.Status)
	}
}
