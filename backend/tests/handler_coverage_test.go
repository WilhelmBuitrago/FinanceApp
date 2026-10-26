package tests

import (
	"net/http"
	"strings"
	"testing"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

// The route groups below ran through the real router but had no test of
// their own: every one of them is reachable by the desktop client. Each
// test seeds the minimum state and pins the contract a client relies on.

func TestAccountsRoutes_CRUDReadPaths(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	code, body := hydrationDo(t, router, http.MethodPost, "/accounts", `{"id":"acc-cov","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","status":"ACTIVE","institution":"NU"}`)
	if code != http.StatusCreated {
		t.Fatalf("create account: %d %s", code, body)
	}
	if code, body = hydrationDo(t, router, http.MethodGet, "/accounts", ""); code != http.StatusOK || !strings.Contains(body, "acc-cov") {
		t.Errorf("list accounts must carry the created account: %d %s", code, body)
	}
	if code, body = hydrationDo(t, router, http.MethodGet, "/accounts/acc-cov", ""); code != http.StatusOK || !strings.Contains(body, "Nu") {
		t.Errorf("get account must carry the account: %d %s", code, body)
	}
}

func TestCategoriesRoutes_FullCycle(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	code, body := hydrationDo(t, router, http.MethodPost, "/categories", `{"id":"cat-cov","name":"Mercado","category_type":"EXPENSE"}`)
	if code != http.StatusCreated {
		t.Fatalf("create category: %d %s", code, body)
	}
	if code, body = hydrationDo(t, router, http.MethodGet, "/categories", ""); code != http.StatusOK || !strings.Contains(body, "cat-cov") {
		t.Errorf("list categories must carry the created one: %d %s", code, body)
	}
	if code, body = hydrationDo(t, router, http.MethodPut, "/categories/cat-cov", `{"name":"Mercado grande","category_type":"EXPENSE"}`); code != http.StatusOK || !strings.Contains(body, "Mercado") {
		t.Errorf("update category must echo the category: %d %s", code, body)
	}
	if code, _ = hydrationDo(t, router, http.MethodDelete, "/categories/cat-cov", ""); code != http.StatusNoContent {
		t.Errorf("delete category: %d", code)
	}
	if code, body = hydrationDo(t, router, http.MethodGet, "/categories", ""); code != http.StatusOK || strings.Contains(body, "cat-cov") {
		t.Errorf("deleted category must be gone from the list: %d %s", code, body)
	}
	if code, _ = hydrationDo(t, router, http.MethodDelete, "/categories/cat-missing", ""); code != http.StatusNotFound {
		t.Errorf("delete unknown category must be 404, got %d", code)
	}
}

func TestDashboardRoutes_ShapeAndRangeFilter(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	if code, body := hydrationDo(t, router, http.MethodPost, "/accounts", `{"id":"acc-dash","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","institution":"NU"}`); code != http.StatusCreated {
		t.Fatalf("setup account: %d %s", code, body)
	}
	for _, mv := range []string{
		`{"account_id":"acc-dash","amount_minor":9000,"type":"INCOME","date":"2026-10-05","description":"sueldo"}`,
		`{"account_id":"acc-dash","amount_minor":-1500,"type":"EXPENSE","date":"2026-10-06","description":"mercado"}`,
	} {
		if code, body := hydrationDo(t, router, http.MethodPost, "/movements", mv); code != http.StatusCreated {
			t.Fatalf("setup movement: %d %s", code, body)
		}
	}

	code, body := hydrationDo(t, router, http.MethodGet, "/dashboard/cash-flow?from=2026-10-01&to=2026-10-31", "")
	if code != http.StatusOK {
		t.Fatalf("cash flow: %d %s", code, body)
	}
	for _, key := range []string{`"inflow"`, `"outflow"`, `"net"`} {
		if !strings.Contains(body, key) {
			t.Errorf("cash flow must carry %s: %s", key, body)
		}
	}
	if !strings.Contains(body, "9000") || !strings.Contains(body, "1500") {
		t.Errorf("cash flow must aggregate both movements: %s", body)
	}

	// The range filter excludes everything outside it.
	code, body = hydrationDo(t, router, http.MethodGet, "/dashboard/cash-flow?from=2027-01-01&to=2027-01-31", "")
	if code != http.StatusOK {
		t.Fatalf("cash flow with exclusive range: %d %s", code, body)
	}
	// Pinned as-is: an empty aggregation encodes as null (the three dashboard
	// routes share this shape). Flagged as a residual, not changed here.
	if got := strings.TrimSpace(body); got != "[]" && got != "null" {
		t.Errorf("an exclusive range must return no points, got: %s", got)
	}

	for _, path := range []string{"/dashboard/income-distribution", "/dashboard/expense-distribution"} {
		code, body := hydrationDo(t, router, http.MethodGet, path, "")
		if code != http.StatusOK {
			t.Errorf("%s: %d %s", path, code, body)
			continue
		}
		if !strings.Contains(body, "category") || !strings.Contains(body, "amount_minor") {
			t.Errorf("%s must carry category and amount buckets: %s", path, body)
		}
	}
}

func TestDocumentsRoutes_ReadPaths(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	repoDB := &repositories.DB{Querier: db.New(sqlDB)}
	if _, err := repositories.NewDocumentRepository(repoDB).Create(t.Context(), domain.Document{
		ID:           "doc-cov",
		DocumentType: "PDF",
		Institution:  strPtr("NU"),
		Status:       "PROCESSED",
		CreatedAt:    "2026-10-20T00:00:00Z",
		UpdatedAt:    "2026-10-20T00:00:00Z",
	}); err != nil {
		t.Fatalf("seed document: %v", err)
	}

	if code, body := hydrationDo(t, router, http.MethodGet, "/documents", ""); code != http.StatusOK || !strings.Contains(body, "doc-cov") {
		t.Errorf("list documents: %d %s", code, body)
	}
	if code, body := hydrationDo(t, router, http.MethodGet, "/documents/recent", ""); code != http.StatusOK || !strings.Contains(body, "doc-cov") {
		t.Errorf("recent documents: %d %s", code, body)
	}
	if code, body := hydrationDo(t, router, http.MethodGet, "/documents/doc-cov", ""); code != http.StatusOK || !strings.Contains(body, "doc-cov") {
		t.Errorf("get document: %d %s", code, body)
	}
	if code, _ := hydrationDo(t, router, http.MethodGet, "/documents/doc-nope", ""); code != http.StatusNotFound {
		t.Errorf("get unknown document must be 404, got %d", code)
	}
}

func TestJobsRoutes_Contract(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	repoDB := &repositories.DB{Querier: db.New(sqlDB)}
	if _, err := repositories.NewImportJobRepo(repoDB).Create(t.Context(), domain.ImportJob{
		ID:         "job-cov",
		DocumentID: "doc-cov",
		Status:     "PROCESSING",
		CreatedAt:  "2026-10-20T00:00:00Z",
	}); err != nil {
		t.Fatalf("seed job: %v", err)
	}

	if code, body := hydrationDo(t, router, http.MethodGet, "/jobs/job-cov", ""); code != http.StatusOK || !strings.Contains(body, "job-cov") {
		t.Errorf("get job by path: %d %s", code, body)
	}
	// A job whose processing has not produced artifacts yet answers 400,
	// never a fabricated empty artifact.
	if code, body := hydrationDo(t, router, http.MethodGet, "/jobs/job-cov/parsed", ""); code != http.StatusBadRequest || !strings.Contains(body, "processing not ready") {
		t.Errorf("parsed without processing: %d %s", code, body)
	}
	if code, body := hydrationDo(t, router, http.MethodGet, "/jobs/job-cov/manifest", ""); code != http.StatusBadRequest || !strings.Contains(body, "processing not ready") {
		t.Errorf("manifest without processing: %d %s", code, body)
	}
	if code, _ := hydrationDo(t, router, http.MethodGet, "/jobs/job-nope/parsed", ""); code != http.StatusNotFound {
		t.Errorf("parsed of unknown job must be 404, got %d", code)
	}
}

// TestAdminPurge_EmptiesTheDatabase pins the destructive maintenance route
// end to end: it opens its own connection from DATABASE_PATH (the harness
// points that at the temporary database), so the wipe is real and the
// assertion is that the seeded account is gone afterwards.
func TestAdminPurge_EmptiesTheDatabase(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	if code, body := hydrationDo(t, router, http.MethodPost, "/accounts", `{"id":"acc-purge","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","institution":"NU"}`); code != http.StatusCreated {
		t.Fatalf("setup account: %d %s", code, body)
	}
	if code, body := hydrationDo(t, router, http.MethodDelete, "/admin/purge", ""); code != http.StatusNoContent {
		t.Fatalf("purge: %d %s", code, body)
	}
	code, body := hydrationDo(t, router, http.MethodGet, "/accounts", "")
	if code != http.StatusOK {
		t.Fatalf("list after purge: %d %s", code, body)
	}
	if strings.Contains(body, "acc-purge") {
		t.Errorf("purge must leave no account behind: %s", body)
	}
}
