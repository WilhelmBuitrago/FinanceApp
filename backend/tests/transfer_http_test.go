package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

/* The transfer contract over HTTP: POST /movements with type TRANSFER
   and a destination creates both legs; the distributions stay neutral;
   deleting a leg deletes the transfer; the legacy single-leg transfer
   (no destination) keeps working. */

type leg struct {
	ID          string `json:"id"`
	AccountID   string `json:"AccountID"`
	AmountMinor int64  `json:"AmountMinor"`
}

func postTransfer(t *testing.T, router http.Handler, body string) (int, string) {
	t.Helper()
	return hydrationDo(t, router, http.MethodPost, "/movements", body)
}

func getBody(t *testing.T, router http.Handler, path string) (int, string) {
	t.Helper()
	return hydrationDo(t, router, http.MethodGet, path, "")
}

func TestTransferEndpoint_CreatesBothLegsUnderOneEvent(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	for _, acc := range []string{"acc-t1", "acc-t2"} {
		if code, body := hydrationDo(t, router, http.MethodPost, "/accounts", `{"id":"`+acc+`","user_id":"u","name":"`+acc+`","account_type":"BANK_ACCOUNT","currency":"COP","institution":"NU"}`); code != http.StatusCreated {
			t.Fatalf("setup account %s: %d %s", acc, code, body)
		}
	}

	code, body := postTransfer(t, router, `{"account_id":"acc-t1","to_account_id":"acc-t2","amount_minor":250000,"type":"TRANSFER","date":"2026-10-26","description":"Alquiler"}`)
	if code != http.StatusCreated {
		t.Fatalf("transfer: %d %s", code, body)
	}
	var created struct {
		ID            string `json:"id"`
		CounterpartID string `json:"counterpart_id"`
	}
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatalf("decode: %v (%s)", err, body)
	}
	if created.ID == "" || created.CounterpartID == "" {
		t.Fatalf("the response must carry both legs: %s", body)
	}

	code, body = getBody(t, router, "/movements")
	if code != http.StatusOK {
		t.Fatalf("list movements: %d %s", code, body)
	}
	var legs []leg
	if err := json.Unmarshal([]byte(body), &legs); err != nil {
		t.Fatalf("decode movements: %v (%s)", err, body)
	}
	if len(legs) != 2 {
		t.Fatalf("a transfer must create exactly two legs, got %d: %s", len(legs), body)
	}
	for _, l := range legs {
		want := int64(-250000)
		if l.AccountID == "acc-t2" {
			want = 250000
		}
		if l.AccountID != "acc-t1" && l.AccountID != "acc-t2" {
			t.Errorf("unexpected account %s", l.AccountID)
		}
		if l.AmountMinor != want {
			t.Errorf("leg in %s must carry %d, got %d", l.AccountID, want, l.AmountMinor)
		}
	}
}

func TestTransferEndpoint_StaysOutOfTheDistributions(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	for _, acc := range []string{"acc-d1", "acc-d2"} {
		if code, body := hydrationDo(t, router, http.MethodPost, "/accounts", `{"id":"`+acc+`","user_id":"u","name":"`+acc+`","account_type":"BANK_ACCOUNT","currency":"COP","institution":"NU"}`); code != http.StatusCreated {
			t.Fatalf("setup account %s: %d %s", acc, code, body)
		}
	}
	if code, body := postTransfer(t, router, `{"account_id":"acc-d1","to_account_id":"acc-d2","amount_minor":250000,"type":"TRANSFER","date":"2026-10-26","description":"Alquiler"}`); code != http.StatusCreated {
		t.Fatalf("transfer: %d %s", code, body)
	}

	for _, path := range []string{"/dashboard/income-distribution", "/dashboard/expense-distribution"} {
		code, body := getBody(t, router, path)
		if code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, code, body)
		}
		if strings.Contains(body, "2.500") {
			t.Errorf("a transfer must not count as income or expense: %s -> %s", path, body)
		}
	}
}

func TestTransferEndpoint_RejectsBadRequestsWithoutWritingAnything(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"same account on both sides", `{"account_id":"acc-x","to_account_id":"acc-x","amount_minor":250000,"type":"TRANSFER","date":"2026-10-26"}`},
		{"unknown destination", `{"account_id":"acc-x","to_account_id":"acc-nope","amount_minor":250000,"type":"TRANSFER","date":"2026-10-26"}`},
		{"zero amount", `{"account_id":"acc-x","to_account_id":"acc-y","amount_minor":0,"type":"TRANSFER","date":"2026-10-26"}`},
		{"negative amount", `{"account_id":"acc-x","to_account_id":"acc-y","amount_minor":-250000,"type":"TRANSFER","date":"2026-10-26"}`},
	}

	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()
	for _, acc := range []string{"acc-x", "acc-y"} {
		if code, body := hydrationDo(t, router, http.MethodPost, "/accounts", `{"id":"`+acc+`","user_id":"u","name":"`+acc+`","account_type":"BANK_ACCOUNT","currency":"COP","institution":"NU"}`); code != http.StatusCreated {
			t.Fatalf("setup account %s: %d %s", acc, code, body)
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, body := postTransfer(t, router, tc.body)
			if code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", code, body)
			}
		})
	}

	// Nothing of any rejected transfer may have been written.
	code, body := getBody(t, router, "/movements")
	if code != http.StatusOK {
		t.Fatalf("list movements: %d %s", code, body)
	}
	if strings.TrimSpace(body) != "[]" && strings.TrimSpace(body) != "null" {
		t.Errorf("no movement may survive a rejected transfer: %s", body)
	}
}

func TestTransferEndpoint_DeletingALegDeletesTheTransfer(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	for _, acc := range []string{"acc-c1", "acc-c2"} {
		if code, body := hydrationDo(t, router, http.MethodPost, "/accounts", `{"id":"`+acc+`","user_id":"u","name":"`+acc+`","account_type":"BANK_ACCOUNT","currency":"COP","institution":"NU"}`); code != http.StatusCreated {
			t.Fatalf("setup account %s: %d %s", acc, code, body)
		}
	}
	code, body := postTransfer(t, router, `{"account_id":"acc-c1","to_account_id":"acc-c2","amount_minor":250000,"type":"TRANSFER","date":"2026-10-26","description":"Alquiler"}`)
	if code != http.StatusCreated {
		t.Fatalf("transfer: %d %s", code, body)
	}
	var created struct {
		ID            string `json:"id"`
		CounterpartID string `json:"counterpart_id"`
	}
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatalf("decode: %v (%s)", err, body)
	}

	code, body = hydrationDo(t, router, http.MethodDelete, "/movements/"+created.ID, "")
	if code != http.StatusNoContent {
		t.Fatalf("delete a leg: %d %s", code, body)
	}

	code, body = getBody(t, router, "/movements")
	if code != http.StatusOK {
		t.Fatalf("list movements: %d %s", code, body)
	}
	if strings.TrimSpace(body) != "[]" && strings.TrimSpace(body) != "null" {
		t.Errorf("deleting a leg must delete the whole transfer: %s", body)
	}
}

func TestTransferEndpoint_LegacySingleLegStillWorks(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	if code, body := hydrationDo(t, router, http.MethodPost, "/accounts", `{"id":"acc-l1","user_id":"u","name":"acc-l1","account_type":"BANK_ACCOUNT","currency":"COP","institution":"NU"}`); code != http.StatusCreated {
		t.Fatalf("setup account: %d %s", code, body)
	}
	// A TRANSFER without destination keeps the old contract: one leg, no
	// relation (the shape old clients send).
	code, body := postTransfer(t, router, `{"account_id":"acc-l1","amount_minor":250000,"type":"TRANSFER","date":"2026-10-26","description":"Alquiler"}`)
	if code != http.StatusCreated {
		t.Fatalf("legacy transfer: %d %s", code, body)
	}
	code, body = getBody(t, router, "/movements")
	if code != http.StatusOK {
		t.Fatalf("list movements: %d %s", code, body)
	}
	var legs []leg
	if err := json.Unmarshal([]byte(body), &legs); err != nil {
		t.Fatalf("decode: %v (%s)", err, body)
	}
	if len(legs) != 1 {
		t.Errorf("a transfer without destination must create a single leg, got %d", len(legs))
	}
}
