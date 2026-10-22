package tests

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"financeapp/backend/internal/app"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/domain/entities"
	"financeapp/backend/internal/repositories"
	_ "modernc.org/sqlite"
)

// setupReviewScenario leaves an account, a document, a job and one parked
// review entry produced by the import, ready to be resolved.
func setupReviewScenario(t *testing.T) (*sql.DB, string) {
	t.Helper()
	artifacts := t.TempDir()
	t.Setenv("ARTIFACTS_PATH", artifacts)

	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	persistence := repositories.NewPersistence(sqlDB)
	now := time.Now().UTC().Format(time.RFC3339)
	ctx := context.Background()

	if _, err := persistence.AccountRepo.Create(ctx, domain.Account{
		ID: "acc-rev", UserID: "u", Name: "Nu", AccountType: "BANK_ACCOUNT",
		Currency: "COP", Status: "ACTIVE", Institution: strPtr("NU"),
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("account: %v", err)
	}
	if _, err := persistence.DocumentRepo.Create(ctx, domain.Document{
		ID: "doc-rev", AccountID: strPtr("acc-rev"), DocumentType: "STATEMENT",
		Status: "UPLOADED", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("document: %v", err)
	}

	jobRepo := repositories.NewImportJobRepo(&repositories.DB{Querier: db.New(sqlDB)})
	procID := "proc-rev"
	if _, err := jobRepo.Create(ctx, entities.ImportJob{
		ID: "job-rev", DocumentID: "doc-rev", Status: "PROCESSING",
		CreatedAt: now, ProcessingID: strPtr(procID),
	}); err != nil {
		t.Fatalf("job: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(artifacts, procID), 0755); err != nil {
		t.Fatal(err)
	}
	parsed := `{
		"movements": [],
		"review_required": [
			{"date_original": "12", "description_original": "almacen dudoso", "amount_original": "+$--", "page_number": 2, "reason": "unparseable amount"}
		],
		"validation": {"valid": true}
	}`
	if err := os.WriteFile(filepath.Join(artifacts, procID, "parsed.json"), []byte(parsed), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(artifacts, procID, "manifest.json"), []byte(`{"validation":{"valid":true}}`), 0644); err != nil {
		t.Fatal(err)
	}

	service := app.NewImportService(config.Load(), persistence, jobRepo)
	if err := service.Import(ctx, "job-rev"); err != nil {
		t.Fatalf("import with a parked movement failed: %v", err)
	}
	return sqlDB, "doc-rev_rev_1"
}

func TestReviewService_Resolve_CreatesMovementAndMarksEntry(t *testing.T) {
	sqlDB, entryID := setupReviewScenario(t)
	defer sqlDB.Close()
	ctx := context.Background()

	service := app.NewReviewService(sqlDB)
	created, err := service.Resolve(ctx, entryID, "Almacén Éxito verificado", -55000, "2026-10-12", "EXPENSE", "acc-rev", nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if created.AmountMinor != -55000 {
		t.Fatalf("wrong amount: %d", created.AmountMinor)
	}

	// The entry is resolved with the correction recorded.
	var reviewStatus, correctedDesc string
	var correctedAmount int64
	if err := sqlDB.QueryRow(
		"SELECT review_status, corrected_description, corrected_amount_minor FROM source_entries WHERE id=?",
		entryID,
	).Scan(&reviewStatus, &correctedDesc, &correctedAmount); err != nil {
		t.Fatal(err)
	}
	if reviewStatus != "RESOLVED" {
		t.Fatalf("entry not resolved: %q", reviewStatus)
	}
	if correctedDesc != "Almacén Éxito verificado" || correctedAmount != -55000 {
		t.Fatalf("correction not recorded: %q / %d", correctedDesc, correctedAmount)
	}

	// The queue is empty and the verified movement exists.
	entries, err := service.ListEntries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty queue, got %d", len(entries))
	}
	var movements int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM movements").Scan(&movements); err != nil {
		t.Fatal(err)
	}
	if movements != 1 {
		t.Fatalf("expected 1 movement after resolve, got %d", movements)
	}
}

func TestReviewService_Resolve_TwiceIsRejected(t *testing.T) {
	sqlDB, entryID := setupReviewScenario(t)
	defer sqlDB.Close()

	service := app.NewReviewService(sqlDB)
	if _, err := service.Resolve(context.Background(), entryID, "primera", -1000, "2026-10-12", "EXPENSE", "acc-rev", nil); err != nil {
		t.Fatalf("first resolve: %v", err)
	}
	if _, err := service.Resolve(context.Background(), entryID, "segunda", -2000, "2026-10-12", "EXPENSE", "acc-rev", nil); err == nil {
		t.Fatal("expected the second resolve to fail: the entry is no longer pending")
	}

	// Only the first correction exists.
	var movements, resolved int
	sqlDB.QueryRow("SELECT COUNT(*) FROM movements").Scan(&movements)
	sqlDB.QueryRow("SELECT COUNT(*) FROM source_entries WHERE review_status='RESOLVED'").Scan(&resolved)
	if movements != 1 || resolved != 1 {
		t.Fatalf("double submission leaked data: movements=%d resolved=%d", movements, resolved)
	}
}
