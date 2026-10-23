package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"financeapp/backend/internal/app"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/repositories"
)

// TestResolvePeriod_DerivesParkedMovements pins L2: confirming the period
// with month+year updates the document, derives the parked movements of
// THAT document against the chosen month and imports the ones that fit;
// the ones that do not fit stay parked; imported movements never move.
func TestResolvePeriod_DerivesParkedMovements(t *testing.T) {
	service, sqlDB := newImportTestEnv(t)
	defer func() { _ = sqlDB.Close() }()
	reviews := app.NewReviewService(sqlDB)

	processingID := "proc-pr-1"
	docID, jobID := seedDocAndJob(t, sqlDB, processingID)

	dir := filepath.Join(os.Getenv("ARTIFACTS_PATH"), processingID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir artifacts: %v", err)
	}
	parsed := `{
      "metadata": {},
      "summary": {},
      "movements": [
        {"date_original": "10 abr 2026", "date_normalized": "2026-04-10", "description_original": "sano", "amount_original": "-$3.000,00", "amount_normalized": -3000, "page_number": 1, "sequence": 1}
      ],
      "review_required": [
        {"date_original": "14 abr", "description_original": "encaja", "amount_original": "+$16.500,00", "amount_normalized": 16500, "page_number": 1, "reason": "unparseable date: Unable to parse period"},
        {"date_original": "15 may", "description_original": "no encaja", "amount_original": "-$9.900,00", "amount_normalized": -9900, "page_number": 1, "reason": "unparseable date: Unable to parse period"}
      ],
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

	var itemID string
	if err := sqlDB.QueryRow(`SELECT id FROM review_items WHERE kind = 'DOCUMENT_PERIOD'`).Scan(&itemID); err != nil {
		t.Fatalf("read document review item: %v", err)
	}

	// Resolve the document period: month ABR (4), year 2026.
	if err := reviews.ResolvePeriod(t.Context(), itemID, 4, 2026); err != nil {
		t.Fatalf("resolve period: %v", err)
	}

	var start, end string
	if err := sqlDB.QueryRow(`SELECT period_start, period_end FROM documents WHERE id = ?`, docID).Scan(&start, &end); err != nil {
		t.Fatalf("read document period: %v", err)
	}
	if start != "2026-04-01" || end != "2026-04-30" {
		t.Fatalf("document period must be the full chosen month, got %s..%s", start, end)
	}

	var itemStatus string
	if err := sqlDB.QueryRow(`SELECT status FROM review_items WHERE id = ?`, itemID).Scan(&itemStatus); err != nil {
		t.Fatalf("read item: %v", err)
	}
	if itemStatus != "RESOLVED" {
		t.Fatalf("document review item must close RESOLVED, got %s", itemStatus)
	}

	// The fitting movement was derived and imported; the other one stays parked.
	var movements int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM movements`).Scan(&movements); err != nil {
		t.Fatalf("count movements: %v", err)
	}
	if movements != 2 {
		t.Fatalf("expected 2 movements (1 original + 1 derived), got %d", movements)
	}
	var derivedAmount int64
	if err := sqlDB.QueryRow(`SELECT amount_minor FROM movements WHERE amount_minor = 16500`).Scan(&derivedAmount); err != nil {
		t.Fatalf("derived movement missing: %v", err)
	}

	var resolvedEntries int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM source_entries WHERE review_status = 'RESOLVED'`).Scan(&resolvedEntries); err != nil {
		t.Fatalf("count resolved entries: %v", err)
	}
	if resolvedEntries != 1 {
		t.Fatalf("the fitting parked movement must be resolved, got %d", resolvedEntries)
	}
	var stillParked int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM source_entries WHERE review_status = 'REVIEW_REQUIRED'`).Scan(&stillParked); err != nil {
		t.Fatalf("count parked entries: %v", err)
	}
	if stillParked != 1 {
		t.Fatalf("the non-fitting movement must stay parked, got %d", stillParked)
	}

	// The original movement is untouched.
	var originalDate string
	if err := sqlDB.QueryRow(`SELECT movement_date FROM movements WHERE amount_minor = -3000`).Scan(&originalDate); err != nil {
		t.Fatalf("original movement must survive: %v", err)
	}
	if originalDate != "2026-04-10" {
		t.Fatalf("imported movement date must not change, got %s", originalDate)
	}
}

// TestResolvePeriod_RejectsForeignKind pins that only DOCUMENT_PERIOD items
// are resolved through this door.
func TestResolvePeriod_RejectsForeignKind(t *testing.T) {
	_, sqlDB := newImportTestEnv(t)
	defer func() { _ = sqlDB.Close() }()
	reviews := app.NewReviewService(sqlDB)

	repo := repositories.NewReviewItemsRepository(&repositories.DB{Querier: db.New(sqlDB)})
	now := time.Now().UTC().Format(time.RFC3339)
	if err := repo.Create(t.Context(), "item-x", "EXTRACTION", strPtr(`{"document_id":"d"}`), nil, now); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if err := reviews.ResolvePeriod(t.Context(), "item-x", 4, 2026); err == nil {
		t.Fatal("resolving a non-document item through the period door must fail")
	}
	var status string
	if err := sqlDB.QueryRow(`SELECT status FROM review_items WHERE id = 'item-x'`).Scan(&status); err != nil {
		t.Fatalf("read item: %v", err)
	}
	if status != "PENDING" {
		t.Fatalf("a rejected resolve must not consume the item, got %s", status)
	}
}
