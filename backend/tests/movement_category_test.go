package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"financeapp/backend/internal/api"
	"financeapp/backend/internal/app"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/repositories"
)

// TestMovement_Category_Assignment_Contract pins the movement/category
// contract the desktop detail drawer relies on since v0.1.1: the list
// handlers expose categoryName/categoryId, and PUT /movements/{id}/category
// assigns and clears the category. Clearing (null) is a valid state: a
// movement without a category is honest data, never a review item.
func TestMovement_Category_Assignment_Contract(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "category-test.db")
	// setMovementCategory opens its own sqlite connection on
	// cfg.DatabasePath (raw-SQL handler, documented): point it at the test
	// database so it can never touch the real one.
	t.Setenv("DATABASE_PATH", dbPath)
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Close before the test's TempDir cleanup: on Windows an open sqlite
	// handle blocks the directory removal.
	defer func() { _ = sqlDB.Close() }()
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
		app.NewMovementService(sqlDB, app.NewReviewService(sqlDB)),
		app.NewReviewService(sqlDB),
		nil,
		nil)

	post := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr
	}
	put := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader([]byte(body)))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr
	}
	getMovements := func() []struct {
		ID           string `json:"ID"`
		CategoryName string `json:"categoryName"`
		CategoryID   string `json:"categoryId"`
	} {
		req := httptest.NewRequest(http.MethodGet, "/movements", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET /movements: expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
		var out []struct {
			ID           string `json:"ID"`
			CategoryName string `json:"categoryName"`
			CategoryID   string `json:"categoryId"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode /movements: %v", err)
		}
		return out
	}

	// Seed: account, category, movement (without a category).
	if rr := post("/accounts", `{"id":"acc-cat","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","status":"ACTIVE","institution":"NU"}`); rr.Code != http.StatusCreated {
		t.Fatalf("seed account: %d %s", rr.Code, rr.Body.String())
	}
	if rr := post("/categories", `{"id":"cat-cat-1","name":"Supermercado","category_type":"EXPENSE"}`); rr.Code != http.StatusCreated {
		t.Fatalf("seed category: %d %s", rr.Code, rr.Body.String())
	}
	mvRR := post("/movements?force=true", `{"account_id":"acc-cat","amount_minor":-2500,"type":"EXPENSE","date":"2026-10-01","description":"Compra"}`)
	if mvRR.Code != http.StatusCreated {
		t.Fatalf("seed movement: %d %s", mvRR.Code, mvRR.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(mvRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created movement: %v", err)
	}
	if created.ID == "" {
		t.Fatal("created movement returned no id")
	}

	// The DTO carries the category fields even when the movement has none.
	movements := getMovements()
	if len(movements) != 1 {
		t.Fatalf("expected 1 movement, got %d", len(movements))
	}
	if movements[0].ID != created.ID {
		t.Fatalf("movement id mismatch: %s", movements[0].ID)
	}
	if movements[0].CategoryName != "" {
		t.Fatalf("uncategorized movement must expose an empty categoryName, got %q", movements[0].CategoryName)
	}

	// Assign: the event behind the movement gains the category.
	if rr := put("/movements/"+created.ID+"/category", `{"categoryId":"cat-cat-1"}`); rr.Code != http.StatusNoContent {
		t.Fatalf("PUT category: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
	movements = getMovements()
	if movements[0].CategoryName != "Supermercado" {
		t.Fatalf("expected categoryName Supermercado, got %q", movements[0].CategoryName)
	}
	if movements[0].CategoryID != "cat-cat-1" {
		t.Fatalf("expected categoryId cat-cat-1, got %q", movements[0].CategoryID)
	}

	// Clear: null leaves the movement uncategorized, a valid state.
	if rr := put("/movements/"+created.ID+"/category", `{"categoryId":null}`); rr.Code != http.StatusNoContent {
		t.Fatalf("PUT null category: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
	movements = getMovements()
	if movements[0].CategoryName != "" || movements[0].CategoryID != "" {
		t.Fatalf("cleared category must leave the movement uncategorized, got name=%q id=%q", movements[0].CategoryName, movements[0].CategoryID)
	}
}
