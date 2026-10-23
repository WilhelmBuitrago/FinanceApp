package tests

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"financeapp/backend/internal/app"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/domain/entities"
	"financeapp/backend/internal/repositories"
)

// newImportTestEnv wires a real migrated database, real repositories and an
// ImportService over isolated paths.
func newImportTestEnv(t *testing.T) (*app.ImportService, *sql.DB) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("ARTIFACTS_PATH", filepath.Join(tmp, "artifacts"))
	sqlDB, err := sql.Open("sqlite", filepath.Join(tmp, "import-review.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	repoDB := &repositories.DB{Querier: db.New(sqlDB)}
	service := app.NewImportService(config.Load(),
		&repositories.Persistence{DB: sqlDB, Querier: db.New(sqlDB)},
		repositories.NewImportJobRepo(repoDB))
	return service, sqlDB
}

func seedDocAndJob(t *testing.T, sqlDB *sql.DB, processingID string) (docID, jobID string) {
	t.Helper()
	repoDB := &repositories.DB{Querier: db.New(sqlDB)}
	accRepo := repositories.NewAccountRepository(repoDB)
	docRepo := repositories.NewDocumentRepository(repoDB)
	jobRepo := repositories.NewImportJobRepo(&repositories.DB{Querier: db.New(sqlDB)})

	account := domain.Account{
		ID: "acc-dp", UserID: "u", Name: "Nu", AccountType: "BANK_ACCOUNT",
		Currency: "COP", Status: "ACTIVE", Institution: strPtr("NU"),
		CreatedAt: "2026-10-23T00:00:00Z", UpdatedAt: "2026-10-23T00:00:00Z",
	}
	if _, err := accRepo.Create(t.Context(), account); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	doc := domain.Document{
		ID: "doc-dp", AccountID: strPtr("acc-dp"), DocumentType: "PDF",
		Institution: strPtr("NU"), FileHash: strPtr("sha"), SourceUri: strPtr("/tmp/x.pdf"),
		Status: "PROCESSED", CreatedAt: "2026-10-23T00:00:00Z", UpdatedAt: "2026-10-23T00:00:00Z",
	}
	if _, err := docRepo.Create(t.Context(), doc); err != nil {
		t.Fatalf("seed document: %v", err)
	}
	job := entities.ImportJob{
		ID:           "job-dp",
		DocumentID:   doc.ID,
		Status:       "PROCESSING",
		ProcessingID: &processingID,
		CreatedAt:    "2026-10-23T00:00:00Z",
	}
	if _, err := jobRepo.Create(t.Context(), job); err != nil {
		t.Fatalf("seed job: %v", err)
	}
	return doc.ID, job.ID
}

// TestImport_ParksDocumentPeriodReview pins L1 on the backend side: when
// the parsed artifact carries a document_review entry, the import parks one
// DOCUMENT_PERIOD item (with the raw and the account) AND imports the
// movements that parsed fine.
func TestImport_ParksDocumentPeriodReview(t *testing.T) {
	service, sqlDB := newImportTestEnv(t)
	defer func() { _ = sqlDB.Close() }()

	processingID := "proc-dp-1"
	docID, jobID := seedDocAndJob(t, sqlDB, processingID)

	dir := filepath.Join(os.Getenv("ARTIFACTS_PATH"), processingID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir artifacts: %v", err)
	}
	parsed := `{
      "metadata": {"periodo": "no se pudo determinar"},
      "summary": {},
      "movements": [
        {"date_original": "14 abr 2026", "date_normalized": "2026-04-14", "description_original": "recibido", "amount_original": "+$16.500,00", "amount_normalized": 16500, "page_number": 1, "sequence": 1}
      ],
      "review_required": [],
      "document_review": [
        {"raw_period": "no se pudo determinar", "reason": "statement period not parseable from the extraction"}
      ]
    }`
	if err := os.WriteFile(filepath.Join(dir, "parsed.json"), []byte(parsed), 0644); err != nil {
		t.Fatalf("write parsed.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"validation":{"valid":true}}`), 0644); err != nil {
		t.Fatalf("write manifest.json: %v", err)
	}

	if err := service.Import(t.Context(), jobID); err != nil {
		t.Fatalf("import: %v", err)
	}

	var count int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM review_items WHERE kind = 'DOCUMENT_PERIOD'`).Scan(&count); err != nil {
		t.Fatalf("count review items: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 DOCUMENT_PERIOD item, got %d", count)
	}
	var payload string
	if err := sqlDB.QueryRow(`SELECT payload_json FROM review_items WHERE kind = 'DOCUMENT_PERIOD'`).Scan(&payload); err != nil {
		t.Fatalf("read payload: %v", err)
	}
	for _, want := range []string{docID, "acc-dp", "no se pudo determinar"} {
		if !contains(payload, want) {
			t.Fatalf("payload %s must carry %s", payload, want)
		}
	}

	var movements int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM movements`).Scan(&movements); err != nil {
		t.Fatalf("count movements: %v", err)
	}
	if movements != 1 {
		t.Fatalf("the movement that parsed must be imported: expected 1, got %d", movements)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
