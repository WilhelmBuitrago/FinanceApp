package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCreateMovement_ZeroAmountRejected pins that the manual path cannot
// fabricate financial identity: a zero amount is neither income nor
// expense, same criterion as the parser.
func TestCreateMovement_ZeroAmountRejected(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()
	seedNuAccount(t, router)

	body := `{"account_id":"acc-up","amount_minor":0,"type":"EXPENSE","date":"2026-10-23","description":"ajuste"}`
	req := httptest.NewRequest(http.MethodPost, "/movements", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("zero amount: expected %d, got %d: %s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

// TestCreateMovement_UnknownTypeRejected pins the clear 400 instead of a
// schema crash: an unknown event type used to fail as an opaque error.
func TestCreateMovement_UnknownTypeRejected(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()
	seedNuAccount(t, router)

	for _, typ := range []string{"FOO", "ingreso", ""} {
		body := `{"account_id":"acc-up","amount_minor":-1500,"type":"` + typ + `","date":"2026-10-23","description":"x"}`
		req := httptest.NewRequest(http.MethodPost, "/movements", strings.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("type %q: expected %d, got %d: %s", typ, http.StatusBadRequest, rr.Code, rr.Body.String())
		}
	}
}

// TestCreateMovement_TransferStillAccepted pins that the type the UI sends
// for a transfer keeps working: validation tightens the domain, it does not
// shrink it.
func TestCreateMovement_TransferStillAccepted(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()
	seedNuAccount(t, router)

	body := `{"account_id":"acc-up","amount_minor":-1500,"type":"TRANSFER","date":"2026-10-23","description":"transferencia"}`
	req := httptest.NewRequest(http.MethodPost, "/movements", strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("transfer: expected %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}
