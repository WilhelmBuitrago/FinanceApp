package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// hydrationDo executes one request against the router and returns status
// and body.
func hydrationDo(t *testing.T, router http.Handler, method, target, body string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Code, rr.Body.String()
}

// TestMovementListings_CarryTheCategory pins the hydrated listings: every
// route that returns movements carries the category of the financial
// event, and the distributions group by its name. One join replaced two
// lookups per movement; this is the observable contract it must preserve.
func TestMovementListings_CarryTheCategory(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	seed := func(method, target, body string, want int) {
		t.Helper()
		code, resp := hydrationDo(t, router, method, target, body)
		if code != want {
			t.Fatalf("setup %s %s: expected %d, got %d: %s", method, target, want, code, resp)
		}
	}
	seed(http.MethodPost, "/accounts", `{"id":"acc-hyd","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","institution":"NU"}`, http.StatusCreated)
	seed(http.MethodPost, "/categories", `{"id":"cat-hyd","name":"Mercado","category_type":"EXPENSE"}`, http.StatusCreated)
	seed(http.MethodPost, "/movements", `{"account_id":"acc-hyd","amount_minor":-1500,"type":"EXPENSE","date":"2026-10-20","description":"mercado","category_id":"cat-hyd"}`, http.StatusCreated)
	seed(http.MethodPost, "/movements", `{"account_id":"acc-hyd","amount_minor":-700,"type":"EXPENSE","date":"2026-10-21","description":"sin categoria"}`, http.StatusCreated)
	seed(http.MethodPost, "/movements", `{"account_id":"acc-hyd","amount_minor":9000,"type":"INCOME","date":"2026-10-22","description":"sueldo"}`, http.StatusCreated)

	type movementRow struct {
		ID           string `json:"id"`
		CategoryName string `json:"categoryName"`
		CategoryID   string `json:"categoryId"`
	}
	rowsOf := func(target string) []movementRow {
		code, body := hydrationDo(t, router, http.MethodGet, target, "")
		if code != http.StatusOK {
			t.Fatalf("GET %s: %d %s", target, code, body)
		}
		var rows []movementRow
		if err := json.Unmarshal([]byte(body), &rows); err != nil {
			t.Fatalf("GET %s: decode: %v (%s)", target, err, body)
		}
		return rows
	}

	// Every listing carries the category of the categorized movement.
	// /movements/recent only ever exposed categoryName (no id): its DTO is
	// unchanged, so it is asserted by name.
	for _, target := range []string{"/movements", "/accounts/acc-hyd/movements"} {
		found := false
		for _, row := range rowsOf(target) {
			if row.CategoryID != "cat-hyd" {
				continue
			}
			found = true
			if row.CategoryName != "Mercado" {
				t.Errorf("%s: the categorized movement must carry its category name, got %q", target, row.CategoryName)
			}
		}
		if !found {
			t.Errorf("%s: the categorized movement is missing from the listing", target)
		}
	}
	named := 0
	for _, row := range rowsOf("/movements/recent") {
		if row.CategoryName == "Mercado" {
			named++
		}
	}
	if named != 1 {
		t.Errorf("/movements/recent must expose the category name on exactly one row, got %d", named)
	}

	type bucket struct {
		Category    string `json:"category"`
		AmountMinor int64  `json:"amount_minor"`
	}
	buckets := func(target string) map[string]int64 {
		code, body := hydrationDo(t, router, http.MethodGet, target, "")
		if code != http.StatusOK {
			t.Fatalf("GET %s: %d %s", target, code, body)
		}
		var out []bucket
		if err := json.Unmarshal([]byte(body), &out); err != nil {
			t.Fatalf("GET %s: decode: %v (%s)", target, err, body)
		}
		byName := make(map[string]int64, len(out))
		for _, b := range out {
			byName[b.Category] = b.AmountMinor
		}
		return byName
	}

	expense := buckets("/dashboard/expense-distribution")
	if expense["Mercado"] != 1500 {
		t.Errorf("expense distribution must group the categorized expense under its name, got %d", expense["Mercado"])
	}
	if expense["Sin categoría"] != 700 {
		t.Errorf("expense distribution must group the uncategorized expense under 'Sin categoría', got %d", expense["Sin categoría"])
	}
	income := buckets("/dashboard/income-distribution")
	if income["Sin categoría"] != 9000 {
		t.Errorf("income distribution must carry the income amount, got %d", income["Sin categoría"])
	}
	if _, ok := income["Mercado"]; ok {
		t.Error("the expense category must not leak into the income distribution")
	}
}
