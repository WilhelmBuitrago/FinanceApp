package tests

import (
	"testing"

	"financeapp/backend/internal/app"
)

func TestPeriodRangeISO(t *testing.T) {
	cases := []struct {
		year, month int
		start, end  string
	}{
		{2026, 4, "2026-04-01", "2026-04-30"},
		{2024, 2, "2024-02-01", "2024-02-29"}, // leap year
		{2026, 2, "2026-02-01", "2026-02-28"},
		{2026, 12, "2026-12-01", "2026-12-31"},
		{2026, 1, "2026-01-01", "2026-01-31"},
	}
	for _, c := range cases {
		start, end := app.PeriodRangeISO(c.year, c.month)
		if start != c.start || end != c.end {
			t.Fatalf("period %04d-%02d: want %s..%s, got %s..%s", c.year, c.month, c.start, c.end, start, end)
		}
	}
}

func TestDeriveMovementDate(t *testing.T) {
	cases := []struct {
		raw         string
		year, month int
		want        string
		wantOK      bool
	}{
		{"14 abr", 2026, 4, "2026-04-14", true},
		{"10 abr 2026", 2026, 4, "2026-04-10", true}, // decorated with the year
		{"01 mar", 2026, 3, "2026-03-01", true},
		{"15 may", 2026, 4, "", false}, // month outside the period
		{"31 abr", 2026, 4, "", false}, // April has 30 days
		{"29 feb", 2024, 2, "2024-02-29", true},
		{"29 feb", 2026, 2, "", false}, // 2026 is not a leap year
		{"basura", 2026, 4, "", false},
	}
	for _, c := range cases {
		iso, ok := app.DeriveMovementDate(c.raw, c.year, c.month)
		if ok != c.wantOK || (ok && iso != c.want) {
			t.Fatalf("derive %q against %04d-%02d: want %s/%v, got %s/%v", c.raw, c.year, c.month, c.want, c.wantOK, iso, ok)
		}
	}
}

func TestSuggestMonthYear(t *testing.T) {
	cases := []struct {
		name     string
		raws     []string
		wantMon  int
		wantYear int
	}{
		{"majority wins with the year of the dated raw", []string{"14 abr", "10 abr 2026", "03 abr"}, 4, 2026},
		{"no year anywhere suggests month only", []string{"14 abr", "03 abr"}, 4, 0},
		{"tie resolves to the first seen", []string{"15 may", "02 mar"}, 5, 0},
		{"garbage suggests nothing", []string{"basura", "???"}, 0, 0},
		{"empty suggests nothing", nil, 0, 0},
		{"leap date year is suggested", []string{"29 feb 2024", "01 feb"}, 2, 2024},
	}
	for _, c := range cases {
		month, year := app.SuggestMonthYear(c.raws)
		if month != c.wantMon || year != c.wantYear {
			t.Fatalf("%s: want month/year %d/%d, got %d/%d", c.name, c.wantMon, c.wantYear, month, year)
		}
	}
}
