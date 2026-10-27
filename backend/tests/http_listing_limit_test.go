package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

/* The listings cap what they return. Everything used to load the whole
   table into memory on every request: fine at hundreds of rows on one
   account, and the shape of a slow server the day the data grows. The
   default cap must not truncate an ordinary account, and an explicit
   smaller limit must be honored. */

func TestListings_CapAtTheDefaultLimit(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	code, body := hydrationDo(t, router, http.MethodPost, "/accounts", `{"id":"acc-lm","user_id":"u","name":"acc-lm","account_type":"BANK_ACCOUNT","currency":"COP","institution":"NU"}`)
	if code != http.StatusCreated {
		t.Fatalf("setup account: %d %s", code, body)
	}
	// One event, many legs: the listing is what is under test.
	if _, err := sqlDB.Exec(`INSERT INTO financial_events (id, event_type, status, verification_status, created_at, updated_at) VALUES ('fe-lm','EXPENSE','ACTIVE','UNVERIFIED','2026-10-26T00:00:00Z','2026-10-26T00:00:00Z')`); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	for i := 0; i < 505; i++ {
		if _, err := sqlDB.Exec(`INSERT INTO movements (id, financial_event_id, account_id, movement_date, amount_minor, status, created_at, updated_at) VALUES (?, 'fe-lm', 'acc-lm', '2026-10-26', ?, 'ACTIVE', ?, ?)`,
			"mv-lm"+string(rune('a'+i%26))+string(rune('0'+i/26%10))+string(rune('0'+i/260%10)),
			-1000, "2026-10-26T00:00:00Z", "2026-10-26T00:00:00Z"); err != nil {
			t.Fatalf("seed movement %d: %v", i, err)
		}
	}

	var rows []map[string]any
	code, body = hydrationDo(t, router, http.MethodGet, "/movements", "")
	if code != http.StatusOK {
		t.Fatalf("GET /movements: %d %s", code, body)
	}
	if err := json.Unmarshal([]byte(body), &rows); err != nil {
		t.Fatalf("decode: %v (%s)", err, body)
	}
	if len(rows) != 500 {
		t.Errorf("the default listing must cap at 500 rows, got %d", len(rows))
	}

	code, body = hydrationDo(t, router, http.MethodGet, "/movements?limit=7", "")
	if code != http.StatusOK {
		t.Fatalf("GET /movements?limit=7: %d %s", code, body)
	}
	rows = nil
	if err := json.Unmarshal([]byte(body), &rows); err != nil {
		t.Fatalf("decode: %v (%s)", err, body)
	}
	if len(rows) != 7 {
		t.Errorf("an explicit smaller limit must be honored, got %d", len(rows))
	}

	code, body = hydrationDo(t, router, http.MethodGet, "/accounts/acc-lm/movements", "")
	if code != http.StatusOK {
		t.Fatalf("GET account movements: %d %s", code, body)
	}
	rows = nil
	if err := json.Unmarshal([]byte(body), &rows); err != nil {
		t.Fatalf("decode: %v (%s)", err, body)
	}
	if len(rows) != 500 {
		t.Errorf("the account listing must cap too, got %d", len(rows))
	}
}
