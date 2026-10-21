package app

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/domain/entities"
	"financeapp/backend/internal/repositories"
	_ "modernc.org/sqlite"
)

// TestImportService_ReadsConfiguredArtifactsPath pins that the import reads
// the artifacts from the configured path (ARTIFACTS_PATH), never from a
// developer-specific hardcoded directory: the MSI runs on other machines.
func TestImportService_ReadsConfiguredArtifactsPath(t *testing.T) {
	artifacts := t.TempDir()
	t.Setenv("ARTIFACTS_PATH", artifacts)

	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	persistence := repositories.NewPersistence(sqlDB)
	now := time.Now().UTC().Format(time.RFC3339)

	if _, err := persistence.AccountRepo.Create(context.Background(), domain.Account{
		ID: "acc-imp", UserID: "u", Name: "Nu", AccountType: "BANK_ACCOUNT",
		Currency: "COP", Status: "ACTIVE", Institution: strPtr("NU"),
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("account: %v", err)
	}
	if _, err := persistence.DocumentRepo.Create(context.Background(), domain.Document{
		ID: "doc-imp", AccountID: strPtr("acc-imp"), DocumentType: "STATEMENT",
		Status: "UPLOADED", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("document: %v", err)
	}

	jobRepo := repositories.NewImportJobRepo(&repositories.DB{Querier: db.New(sqlDB)})
	procID := "proc-imp"
	if _, err := jobRepo.Create(context.Background(), entities.ImportJob{
		ID: "job-imp", DocumentID: "doc-imp", Status: "PROCESSING",
		CreatedAt: now, ProcessingID: strPtr(procID),
	}); err != nil {
		t.Fatalf("job: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(artifacts, procID), 0755); err != nil {
		t.Fatal(err)
	}
	parsed := `{"validation":{"valid":true},"movements":[{"date_original":"01","date_normalized":"2026-10-01","description_original":"test","amount_normalized":100,"page_number":1,"sequence":1}]}`
	if err := os.WriteFile(filepath.Join(artifacts, procID, "parsed.json"), []byte(parsed), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(artifacts, procID, "manifest.json"), []byte(`{"validation":{"valid":true}}`), 0644); err != nil {
		t.Fatal(err)
	}

	service := NewImportService(config.Load(), persistence, jobRepo)
	if err := service.Import(context.Background(), "job-imp"); err != nil {
		t.Fatalf("import from configured artifacts path failed: %v", err)
	}

	var n int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM movements").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 movement imported, got %d", n)
	}
}
