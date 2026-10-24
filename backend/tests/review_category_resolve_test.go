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

// TestReviewCategoryResolve_MinimalBody fixes the CATEGORY resolution path:
// the frontend sends only {category_id} (the whole human decision), and the
// handler used to reject it with 400 because its validation demanded the
// movement fields (account_id/date/description) that only the movement-
// creating kinds use. The queue is parked by deleting a category, exactly
// like the product does.
func TestReviewCategoryResolve_MinimalBody(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "review-category.db")
	// deleteCategory and setMovementCategory open their own sqlite
	// connections on cfg.DatabasePath (raw-SQL handlers): point it at the
	// test database, never the real one.
	t.Setenv("DATABASE_PATH", dbPath)
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
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
		app.NewMovementService(sqlDB),
		app.NewReviewService(sqlDB),
		nil,
		nil)

	post := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr
	}
	del := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr
	}
	get := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr
	}

	// Seed: account, two categories, two movements on the doomed category.
	if rr := post("/accounts", `{"id":"acc-rc","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","status":"ACTIVE","institution":"NU"}`); rr.Code != http.StatusCreated {
		t.Fatalf("seed account: %d %s", rr.Code, rr.Body.String())
	}
	for _, c := range []string{`{"id":"cat-rc-a","name":"Supermercado","category_type":"EXPENSE"}`, `{"id":"cat-rc-b","name":"Cine","category_type":"EXPENSE"}`} {
		if rr := post("/categories", c); rr.Code != http.StatusCreated {
			t.Fatalf("seed category: %d %s", rr.Code, rr.Body.String())
		}
	}
	if rr := post("/movements?force=true", `{"account_id":"acc-rc","amount_minor":-1000,"type":"EXPENSE","date":"2026-10-01","description":"Mercado","category_id":"cat-rc-a"}`); rr.Code != http.StatusCreated {
		t.Fatalf("seed movement 1: %d %s", rr.Code, rr.Body.String())
	}
	if rr := post("/movements?force=true", `{"account_id":"acc-rc","amount_minor":-2000,"type":"EXPENSE","date":"2026-10-02","description":"Cine","category_id":"cat-rc-a"}`); rr.Code != http.StatusCreated {
		t.Fatalf("seed movement 2: %d %s", rr.Code, rr.Body.String())
	}

	// Deleting the category parks one CATEGORY item per affected movement.
	if rr := del("/categories/cat-rc-a"); rr.Code != http.StatusNoContent {
		t.Fatalf("delete category: %d %s", rr.Code, rr.Body.String())
	}
	queue := get("/review/entries")
	if queue.Code != http.StatusOK {
		t.Fatalf("list queue: %d %s", queue.Code, queue.Body.String())
	}
	var items []struct {
		ID          string `json:"ID"`
		Kind        string `json:"Kind"`
		PayloadJson string `json:"PayloadJson"`
	}
	if err := json.Unmarshal(queue.Body.Bytes(), &items); err != nil {
		t.Fatalf("decode queue: %v", err)
	}
	itemFor := func(movementID string) string {
		for _, it := range items {
			if it.Kind != "CATEGORY" {
				continue
			}
			var payload struct {
				MovementID string `json:"movement_id"`
			}
			if err := json.Unmarshal([]byte(it.PayloadJson), &payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload.MovementID == movementID {
				return it.ID
			}
		}
		t.Fatalf("no CATEGORY item parked for movement %s", movementID)
		return ""
	}

	// Find the two seeded movements to map ids.
	var movements []struct {
		ID           string `json:"ID"`
		Description  string `json:"Description"`
		CategoryName string `json:"categoryName"`
		CategoryID   string `json:"categoryId"`
	}
	decodeMovements := func() {
		rr := get("/movements")
		if rr.Code != http.StatusOK {
			t.Fatalf("list movements: %d %s", rr.Code, rr.Body.String())
		}
		movements = nil
		if err := json.Unmarshal(rr.Body.Bytes(), &movements); err != nil {
			t.Fatalf("decode movements: %v", err)
		}
	}
	decodeMovements()
	byDescription := func(desc string) string {
		for _, m := range movements {
			if m.Description == desc {
				return m.ID
			}
		}
		t.Fatalf("movement %q not found", desc)
		return ""
	}
	mercadoID := byDescription("Mercado")
	cineID := byDescription("Cine")

	// Resolution 1: assign another category, body is ONLY {category_id}.
	rr := post("/review/entries/"+itemFor(mercadoID)+"/resolve", `{"category_id":"cat-rc-b"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("resolve with category: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	// Resolution 2: leave the movement uncategorized (null), the other
	// valid human decision.
	rr = post("/review/entries/"+itemFor(cineID)+"/resolve", `{"category_id":null}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("resolve without category: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// Both decisions landed on the financial event behind the movement.
	decodeMovements()
	for _, m := range movements {
		switch m.ID {
		case mercadoID:
			if m.CategoryName != "Cine" || m.CategoryID != "cat-rc-b" {
				t.Fatalf("expected Mercado reassigned to Cine, got name=%q id=%q", m.CategoryName, m.CategoryID)
			}
		case cineID:
			if m.CategoryName != "" || m.CategoryID != "" {
				t.Fatalf("expected Cine left uncategorized, got name=%q id=%q", m.CategoryName, m.CategoryID)
			}
		}
	}

	// The items are closed: resolving twice must fail (409/400), never
	// silently apply again.
	rr = post("/review/entries/"+itemFor(mercadoID)+"/resolve", `{"category_id":"cat-rc-b"}`)
	if rr.Code == http.StatusCreated {
		t.Fatal("resolving an already-resolved item must not create anything")
	}
}

// TestReviewMovementResolve_StillRequiresCorrectionFields pins the other
// side of the fix: the kinds that CREATE a movement (DUPLICATE here) keep
// demanding the full correction. The per-kind validation must not become a
// blanket relaxation.
func TestReviewMovementResolve_StillRequiresCorrectionFields(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	// Close before the TempDir cleanup: on Windows an open sqlite handle
	// blocks the directory removal.
	defer func() { _ = sqlDB.Close() }()

	post := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr
	}

	if rr := post("/accounts", `{"id":"acc-dr","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","status":"ACTIVE","institution":"NU"}`); rr.Code != http.StatusCreated {
		t.Fatalf("seed account: %d %s", rr.Code, rr.Body.String())
	}
	// Two identical movements: the second collides and parks a DUPLICATE.
	dup := `{"account_id":"acc-dr","amount_minor":-1500,"type":"EXPENSE","date":"2026-10-05","description":"Duplicado"}`
	if rr := post("/movements?force=true", dup); rr.Code != http.StatusCreated {
		t.Fatalf("seed movement: %d %s", rr.Code, rr.Body.String())
	}
	rr := post("/movements", dup)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409 duplicate, got %d: %s", rr.Code, rr.Body.String())
	}
	var parked struct {
		ReviewItemID string `json:"review_item_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &parked); err != nil {
		t.Fatalf("decode 409 body: %v", err)
	}
	if parked.ReviewItemID == "" {
		t.Fatal("duplicate must come back with the parked review item id")
	}

	// Minimal body on a movement kind: still a 400, as designed.
	rr = post("/review/entries/"+parked.ReviewItemID+"/resolve", `{"category_id":null}`)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a movement-kind resolution without correction fields, got %d: %s", rr.Code, rr.Body.String())
	}
	// The full correction resolves it.
	rr = post("/review/entries/"+parked.ReviewItemID+"/resolve", `{"account_id":"acc-dr","amount_minor":-1500,"type":"EXPENSE","date":"2026-10-05","description":"Duplicado confirmado"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("full correction: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}
