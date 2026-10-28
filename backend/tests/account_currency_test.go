package tests

import (
	"net/http"
	"strings"
	"testing"
)

/* The product runs in Colombia for years: one currency. The route
   accepted any string ("cop", "pesos", "COP") as if they were three
   different monies, so the same account could store its amounts under
   one of several spellings. The currency is now normalized to an
   uppercase code from a known catalog, defaulting to COP. */

func TestCreateAccount_NormalizesCurrency(t *testing.T) {
	cases := []struct {
		name     string
		payload  string
		wantCode int
		want     string
	}{
		{"lowercase normalizes", `{"id":"acc-cur1","name":"a","account_type":"BANK_ACCOUNT","institution":"NU","currency":"cop"}`, http.StatusCreated, "COP"},
		{"padded normalizes", `{"id":"acc-cur2","name":"a","account_type":"BANK_ACCOUNT","institution":"NU","currency":"  usd  "}`, http.StatusCreated, "USD"},
		{"absent defaults to COP", `{"id":"acc-cur3","name":"a","account_type":"BANK_ACCOUNT","institution":"NU"}`, http.StatusCreated, "COP"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router, sqlDB := newTestRouter(t)
			defer func() { _ = sqlDB.Close() }()
			code, body := hydrationDo(t, router, http.MethodPost, "/accounts", tc.payload)
			if code != tc.wantCode {
				t.Fatalf("create: %d %s", code, body)
			}
			if !strings.Contains(body, `"Currency":"`+tc.want+`"`) {
				t.Errorf("the stored currency must be %s, got %s", tc.want, body)
			}
		})
	}
}

func TestCreateAccount_RejectsAnUnknownCurrency(t *testing.T) {
	router, sqlDB := newTestRouter(t)
	defer func() { _ = sqlDB.Close() }()

	code, body := hydrationDo(t, router, http.MethodPost, "/accounts", `{"id":"acc-cur-bad","name":"a","account_type":"BANK_ACCOUNT","institution":"NU","currency":"pesos"}`)
	if code != http.StatusBadRequest {
		t.Fatalf("a currency outside the catalog must not be storable, got %d %s", code, body)
	}
	if !strings.Contains(body, "currency") {
		t.Errorf("the error must name the field, got %s", body)
	}
}
