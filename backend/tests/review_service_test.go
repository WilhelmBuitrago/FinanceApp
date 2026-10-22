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

func newReviewTestDB(t *testing.T) (*sql.DB, *repositories.Persistence, string) {
	t.Helper()
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
		ID: "acc-rv", UserID: "u", Name: "Nu", AccountType: "BANK_ACCOUNT",
		Currency: "COP", Status: "ACTIVE", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("account: %v", err)
	}
	if _, err := persistence.CategoryRepo.Create(ctx, domain.Category{
		ID: "cat-rv", Name: "Mercado", CategoryType: strPtr("EXPENSE"),
		Status: "ACTIVE", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("category: %v", err)
	}
	return sqlDB, persistence, now
}

// pendingItemID returns the id of the only pending item of the given kind.
func pendingItemID(t *testing.T, service *app.ReviewService, kind string) string {
	t.Helper()
	items, err := service.ListPending(context.Background())
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	for _, it := range items {
		if it.Kind == kind {
			return it.ID
		}
	}
	t.Fatalf("no pending %s item in queue (%d items)", kind, len(items))
	return ""
}

func TestReviewService_ExtractionResolve_CreatesMovementAndResolves(t *testing.T) {
	artifacts := t.TempDir()
	t.Setenv("ARTIFACTS_PATH", artifacts)

	sqlDB, persistence, now := newReviewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	if _, err := persistence.DocumentRepo.Create(ctx, domain.Document{
		ID: "doc-x", AccountID: strPtr("acc-rv"), DocumentType: "STATEMENT",
		Status: "UPLOADED", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("document: %v", err)
	}
	jobRepo := repositories.NewImportJobRepo(&repositories.DB{Querier: db.New(sqlDB)})
	procID := "proc-x"
	if _, err := jobRepo.Create(ctx, entities.ImportJob{
		ID: "job-x", DocumentID: "doc-x", Status: "PROCESSING",
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

	importService := app.NewImportService(config.Load(), persistence, jobRepo)
	if err := importService.Import(ctx, "job-x"); err != nil {
		t.Fatalf("import: %v", err)
	}

	service := app.NewReviewService(sqlDB)
	itemID := pendingItemID(t, service, "EXTRACTION")

	created, err := service.Resolve(ctx, itemID, "Almacén Éxito verificado", -55000, "2026-10-12", "EXPENSE", "acc-rv", nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if created.AmountMinor != -55000 {
		t.Fatalf("wrong amount: %d", created.AmountMinor)
	}

	// No pending items left and the source entry is resolved too.
	items, _ := service.ListPending(ctx)
	if len(items) != 0 {
		t.Fatalf("queue should be empty, got %d", len(items))
	}
	var reviewStatus string
	if err := sqlDB.QueryRow("SELECT review_status FROM source_entries WHERE review_status != 'VALID'").Scan(&reviewStatus); err != nil {
		t.Fatal(err)
	}
	if reviewStatus != "RESOLVED" {
		t.Fatalf("source entry not resolved: %q", reviewStatus)
	}
}

func TestReviewService_ExtractionDoubleResolveIsRejected(t *testing.T) {
	sqlDB, _, _ := newReviewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	service := app.NewReviewService(sqlDB)
	if _, err := service.Submit(ctx, "EXTRACTION", map[string]any{"reason": "test"}, nil); err != nil {
		t.Fatalf("submit: %v", err)
	}
	itemID := pendingItemID(t, service, "EXTRACTION")

	if _, err := service.Resolve(ctx, itemID, "uno", -1000, "2026-10-01", "EXPENSE", "acc-rv", nil); err != nil {
		t.Fatalf("first resolve: %v", err)
	}
	if _, err := service.Resolve(ctx, itemID, "dos", -2000, "2026-10-01", "EXPENSE", "acc-rv", nil); err == nil {
		t.Fatal("second resolve must fail: the item is no longer pending")
	}
	var movements int
	sqlDB.QueryRow("SELECT COUNT(*) FROM movements").Scan(&movements)
	if movements != 1 {
		t.Fatalf("double submission leaked data: %d movements", movements)
	}
}

func TestReviewService_DuplicateResolve_CreatesTheMovement(t *testing.T) {
	sqlDB, _, _ := newReviewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	service := app.NewReviewService(sqlDB)
	_, err := service.Submit(ctx, "DUPLICATE", map[string]any{
		"account_id": "acc-rv", "amount_minor": -4200, "date": "2026-10-05",
		"type": "EXPENSE", "description": "compra duplicada",
	}, nil)
	if err != nil {
		t.Fatalf("submit duplicate: %v", err)
	}
	itemID := pendingItemID(t, service, "DUPLICATE")

	created, err := service.Resolve(ctx, itemID, "compra duplicada", -4200, "2026-10-05", "EXPENSE", "acc-rv", nil)
	if err != nil {
		t.Fatalf("resolve duplicate: %v", err)
	}
	if created.AmountMinor != -4200 {
		t.Fatalf("wrong amount: %d", created.AmountMinor)
	}
}

func TestReviewService_CategoryResolve_ReassignsTheMovement(t *testing.T) {
	sqlDB, persistence, _ := newReviewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	// A movement under the category that "disappeared".
	fe := domain.FinancialEvent{
		ID: "fe-cat", EventType: "EXPENSE", Status: "ACTIVE",
		VerificationStatus: "VERIFIED", CategoryID: strPtr("cat-rv"),
		Description: strPtr("mercado"), CreatedAt: "2026-10-01T00:00:00Z", UpdatedAt: "2026-10-01T00:00:00Z",
	}
	if _, err := persistence.FinancialEventRepo.Create(ctx, fe); err != nil {
		t.Fatalf("event: %v", err)
	}
	if _, err := persistence.MovementRepo.Create(ctx, domain.Movement{
		ID: "mv-cat", FinancialEventID: "fe-cat", AccountID: "acc-rv",
		MovementDate: "2026-10-01", AmountMinor: -9900, Description: strPtr("mercado"),
		Status: "ACTIVE", CreatedAt: "2026-10-01T00:00:00Z", UpdatedAt: "2026-10-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("movement: %v", err)
	}

	service := app.NewReviewService(sqlDB)
	if _, err := service.Submit(ctx, "CATEGORY", map[string]any{
		"movement_id": "mv-cat", "category_id": "cat-rv", "category_name": "Mercado",
	}, nil); err != nil {
		t.Fatalf("submit category: %v", err)
	}
	itemID := pendingItemID(t, service, "CATEGORY")

	// The human decides: leave it uncategorized.
	if _, err := service.Resolve(ctx, itemID, "", 0, "", "", "", nil); err != nil {
		t.Fatalf("resolve category: %v", err)
	}

	got, err := persistence.FinancialEventRepo.GetByID(ctx, "fe-cat")
	if err != nil {
		t.Fatal(err)
	}
	if got.CategoryID != nil {
		t.Fatalf("category should be cleared, got %v", *got.CategoryID)
	}
}
