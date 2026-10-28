package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"financeapp/backend/internal/api"
	"financeapp/backend/internal/app"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/repositories"
)

func TestReviewDuplicate_ResolveWithoutDescription_Red(t *testing.T) {
	// This test reproduces the bug: resolving a DUPLICATE without description should succeed (201)
	// Currently it returns 400 because handlers_review.go requires req.Description != ""
	// It is RED before fix.
	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil { t.Fatalf("open: %v", err) }
	defer sqlDB.Close()
	if err := db.Apply(sqlDB); err != nil { t.Fatalf("apply: %v", err) }

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
		app.NewTransferService(sqlDB),
		app.NewReviewService(sqlDB),
		nil, nil)

	post := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr
	}

	// seed account
	if rr := post("/accounts", `{"id":"acc1","user_id":"u","name":"A","account_type":"BANK_ACCOUNT","currency":"COP","status":"ACTIVE","institution":"NU"}`); rr.Code != http.StatusCreated {
		t.Fatalf("seed account: %d", rr.Code)
	}
	// create first movement
	if rr := post("/movements?force=true", `{"account_id":"acc1","amount_minor":-1000,"type":"EXPENSE","date":"2026-10-01","description":"X"}`); rr.Code != http.StatusCreated {
		t.Fatalf("seed movement: %d", rr.Code)
	}
	// duplicate -> 409 with review item
	rr := post("/movements", `{"account_id":"acc1","amount_minor":-1000,"type":"EXPENSE","date":"2026-10-01","description":"X"}`)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409 duplicate, got %d %s", rr.Code, rr.Body.String())
	}
	var parked struct { ReviewItemID string `json:"review_item_id"` }
	if err := json.Unmarshal(rr.Body.Bytes(), &parked); err != nil { t.Fatalf("decode: %v", err) }

	// resolve without description (empty string)
	resolveBody := `{"account_id":"acc1","amount_minor":-1000,"type":"EXPENSE","date":"2026-10-01","description":""}`
	rr = post("/review/entries/"+parked.ReviewItemID+"/resolve", resolveBody)
	// AFTER fix we expect 201
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 after fix, got %d body %s", rr.Code, rr.Body.String())
	}
	t.Logf("GREEN: empty description accepted, code %d", rr.Code)
}
