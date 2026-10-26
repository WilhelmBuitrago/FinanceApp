package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestErrorResponsesAreJSONWithTheErrorField pins the API error contract:
// every failure answers application/json with an {"error": ...} body.
//
// http.Error declares text/plain, and the handlers that used it directly
// sent JSON bodies under that wrong content type: a client that trusted
// the header got a parse error instead of the message. The desktop client
// works around it by reading the body as text (api/client.ts); this test
// removes the need for the workaround.
func TestErrorResponsesAreJSONWithTheErrorField(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	cases := []struct {
		name       string
		method     string
		target     string
		body       string
		wantStatus int
	}{
		{"malformed movement body", http.MethodPost, "/movements", `{`, http.StatusBadRequest},
		{"zero amount rejected", http.MethodPost, "/movements", `{"account_id":"a","amount_minor":0,"type":"EXPENSE","date":"2026-10-21"}`, http.StatusBadRequest},
		{"unsupported institution", http.MethodPost, "/accounts", `{"id":"acc-err","institution":"XX","account_type":"BANK_ACCOUNT"}`, http.StatusBadRequest},
		{"unknown category update", http.MethodPut, "/categories/nope", `{"name":"x","category_type":"INCOME"}`, http.StatusNotFound},
		{"unknown document lookup", http.MethodGet, "/documents/nope", "", http.StatusNotFound},
		{"job lookup without processing", http.MethodGet, "/jobs/nope/parsed", "", http.StatusNotFound},
		{"upload without multipart body", http.MethodPost, "/documents/upload", "not a form", http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.body))
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tc.wantStatus, rr.Code, rr.Body.String())
			}
			if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
				t.Errorf("error response must declare application/json, got %q (body %s)", ct, rr.Body.String())
			}
			var payload struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
				t.Fatalf("error body is not JSON: %v (%s)", err, rr.Body.String())
			}
			if payload.Error == "" {
				t.Errorf("error body must carry the error field: %s", rr.Body.String())
			}
		})
	}
}
