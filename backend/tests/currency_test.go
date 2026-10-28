package tests

import (
	"testing"

	"financeapp/backend/internal/domain"
)

func TestNormalizeCurrency_DefaultsTrimsAndValidates(t *testing.T) {
	cases := []struct {
		raw    string
		want   string
		wantOK bool
	}{
		{"", domain.DefaultCurrency, true},
		{"   ", domain.DefaultCurrency, true},
		{"cop", "COP", true},
		{" usd ", "USD", true},
		{"Pesos", "", false},
		{"123", "", false},
		{"copx", "", false},
	}
	for _, tc := range cases {
		got, ok := domain.NormalizeCurrency(tc.raw)
		if got != tc.want || ok != tc.wantOK {
			t.Errorf("NormalizeCurrency(%q) = (%q, %v), want (%q, %v)", tc.raw, got, ok, tc.want, tc.wantOK)
		}
	}
}
