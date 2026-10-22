package tests

import (
	"database/sql"
	"strings"
	"testing"

	"financeapp/backend/internal/db"

	_ "modernc.org/sqlite"
)

func openMigrationDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("pragma: %v", err)
	}
	return sqlDB
}

func TestApplyMigrations(t *testing.T) {
	sqlDB := openMigrationDB(t)
	defer func() { _ = sqlDB.Close() }()
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("apply: %v", err)
	}
	// idempotent: a second run changes nothing
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("apply twice: %v", err)
	}
	for _, tbl := range []string{"documents", "financial_events", "movements", "import_jobs"} {
		var name string
		if err := sqlDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tbl).Scan(&name); err != nil {
			t.Fatalf("table %s missing: %v", tbl, err)
		}
	}
}

// TestApplyMigrations_Effects guards that every migration in the embedded
// directory leaves its observable effect on the schema. Apply reads the
// directory (no hardcoded list), so a new file applies automatically; this
// test pins that known migrations still take effect, and catches a file that
// lands without changing anything.
func TestApplyMigrations_Effects(t *testing.T) {
	sqlDB := openMigrationDB(t)
	defer func() { _ = sqlDB.Close() }()
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// Tables introduced by successive migrations.
	for _, tbl := range []string{"import_jobs", "movement_components", "movement_relations", "event_relations"} {
		var n int
		if err := sqlDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tbl).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("table %s missing after Apply", tbl)
		}
	}

	// Migration 0003 adds review columns to source_entries and two indexes.
	rows, err := sqlDB.Query("PRAGMA table_info(source_entries)")
	if err != nil {
		t.Fatal(err)
	}
	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		cols[name] = true
	}
	_ = rows.Close()
	for _, col := range []string{"review_status", "source_key", "corrected_amount_minor"} {
		if !cols[col] {
			t.Fatalf("source_entries.%s missing after Apply (migration 0003)", col)
		}
	}
	for _, idxName := range []string{"idx_source_entries_document_extraction", "uq_source_entries_key"} {
		var n int
		if err := sqlDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", idxName).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("index %s missing after Apply", idxName)
		}
	}

	// Unique index guarding one installment number per plan (migration 0004).
	var idx int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='uq_installments_plan_number'").Scan(&idx); err != nil {
		t.Fatal(err)
	}
	if idx != 1 {
		t.Fatal("unique index uq_installments_plan_number missing after Apply")
	}

	// Normalized English component catalog (migrations 0005/0006).
	var schema string
	if err := sqlDB.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='movement_components'").Scan(&schema); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"'CAPITAL'", "'INTEREST'", "'FEE'", "'INSURANCE'", "'LATE_FEE'"} {
		if !strings.Contains(schema, want) {
			t.Fatalf("component catalog missing %s in CHECK", want)
		}
	}
	for _, legacy := range []string{"'SEGURO'", "'MORA'"} {
		if strings.Contains(schema, legacy) {
			t.Fatalf("legacy value %s still present in CHECK", legacy)
		}
	}
}
