package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDeleteMovement_Contract pins DELETE /movements after its handler
// stopped opening its own sqlite connection and moved to the repository:
// a happy delete answers 204, deleting an unknown movement is not an
// error (SQLite affects zero rows), and a broken persistence answers 500
// through the repository error taxonomy instead of a fabricated 204.
func TestDeleteMovement_Contract(t *testing.T) {
	router, sqlDB := newTestRouter(t)

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

	if rr := post("/accounts", `{"id":"acc-dm","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","status":"ACTIVE","institution":"NU"}`); rr.Code != http.StatusCreated {
		t.Fatalf("seed account: %d %s", rr.Code, rr.Body.String())
	}
	rr := post("/movements?force=true", `{"account_id":"acc-dm","amount_minor":-1000,"type":"EXPENSE","date":"2026-10-01","description":"Va y vuelve"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("seed movement: %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created movement: %v", err)
	}
	if created.ID == "" {
		t.Fatal("created movement returned no id")
	}

	// The real delete removes the row.
	if rr := del("/movements/" + created.ID); rr.Code != http.StatusNoContent {
		t.Fatalf("delete movement: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
	list := httptest.NewRecorder()
	router.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/movements", nil))
	var movements []struct {
		ID string `json:"ID"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &movements); err != nil {
		t.Fatalf("decode movements: %v", err)
	}
	for _, m := range movements {
		if m.ID == created.ID {
			t.Fatal("deleted movement still listed")
		}
	}

	// Deleting it again is idempotent at the HTTP layer: zero rows, no
	// error, no fabricated detail.
	if rr := del("/movements/" + created.ID); rr.Code != http.StatusNoContent {
		t.Fatalf("second delete: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// A broken persistence must surface as a server error, never as 204.
	_ = sqlDB.Close()
	if rr := del("/movements/ghost"); rr.Code != http.StatusInternalServerError {
		t.Fatalf("delete on broken persistence: expected 500, got %d: %s", rr.Code, rr.Body.String())
	}
}
