package tests

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"financeapp/backend/internal/api"
	"financeapp/backend/internal/app"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/repositories"
)

// newTestRouter builds the real router over a temporary migrated database.
// The returned *sql.DB is the handle to break persistence in failure tests.
func newTestRouter(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "handler-test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	repoDB := &repositories.DB{Querier: db.New(sqlDB)}
	router := api.NewRouter(config.Load(),
		repositories.NewDocumentRepository(repoDB),
		repositories.NewImportJobRepo(&repositories.DB{Querier: db.New(sqlDB)}),
		repositories.NewAccountRepository(repoDB),
		repositories.NewMovementRepository(repoDB),
		repositories.NewFinancialEventRepository(repoDB),
		repositories.NewSourceEntryRepository(repoDB),
		repositories.NewCategoryRepository(repoDB),
		app.NewMovementService(sqlDB),
		app.NewReviewService(sqlDB),
		nil)
	return router, sqlDB
}

// TestPostMovement_PersistenceFailure_ReturnsError pins that a failed write
// is reported as a server error instead of a fabricated 201: the user must
// never see a movement that is not in the database.
func TestPostMovement_PersistenceFailure_ReturnsError(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	// Break persistence underneath the repositories.
	sqlDB.Close()

	body := `{"account_id":"acc-x","amount_minor":-1000,"type":"EXPENSE","date":"2026-10-21","description":"ghost"}`
	req := httptest.NewRequest(http.MethodPost, "/movements?force=true", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d on persistence failure, got %d: %s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
}

// TestPostMovement_Duplicate_StaysConflict pins that the existing duplicate
// detection keeps its 409 contract. The account is created first: without it
// the first POST is a phantom 201 (the write failed and nobody looked).
func TestPostMovement_Duplicate_StaysConflict(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer sqlDB.Close()

	acc := `{"id":"acc-dup","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","status":"ACTIVE","institution":"NU"}`
	req := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(acc))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("setup account: expected %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	first := `{"account_id":"acc-dup","amount_minor":-1500,"type":"EXPENSE","date":"2026-10-20","description":"dup"}`
	req = httptest.NewRequest(http.MethodPost, "/movements?force=true", strings.NewReader(first))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("setup: expected %d on first create, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/movements", strings.NewReader(first))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected %d on duplicate, got %d: %s", http.StatusConflict, rr.Code, rr.Body.String())
	}
	// The duplicate is parked in the unified review queue (not localStorage).
	var reviewID string
	if err := json.Unmarshal(rr.Body.Bytes(), &struct {
		ReviewItemID *string `json:"review_item_id"`
	}{&reviewID}); err != nil {
		t.Fatalf("decode 409 body: %v", err)
	}
	if reviewID == "" {
		t.Fatal("409 must return review_item_id so the UI can track the item")
	}
	qReq := httptest.NewRequest(http.MethodGet, "/review/entries", nil)
	qRR := httptest.NewRecorder()
	router.ServeHTTP(qRR, qReq)
	if qRR.Code != http.StatusOK {
		t.Fatalf("review queue: %d", qRR.Code)
	}
	if !strings.Contains(qRR.Body.String(), reviewID) {
		t.Fatal("parked duplicate not visible in the review queue")
	}
}
