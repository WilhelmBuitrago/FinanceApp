package tests

import (
	"database/sql"
	"path/filepath"
	"testing"

	"financeapp/backend/internal/db"
)

/* The centavo migration: the stored unit. Imported movements arrived in
   whole pesos (the processor's normalize_amount) while the UI, the
   manual path and the schema's name (amount_minor) all mean centavos.
   The migration rescales only the imported rows, and Apply re-runs every
   migration on each startup, so running it twice must not scale twice. */

func TestMigrations_RescaleImportedAmountsToCentavos(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "cents.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	// One imported row (the old, whole-peso convention) and one manual row
	// (always centavos).
	if _, err := sqlDB.Exec(`INSERT INTO financial_events (id, event_type, status, verification_status, created_at, updated_at) VALUES ('fe-cents','EXPENSE','ACTIVE','UNVERIFIED','2026-10-27T00:00:00Z','2026-10-27T00:00:00Z')`); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO accounts (id, user_id, name, account_type, currency, status, created_at, updated_at) VALUES ('acc-cents','u','Nu','BANK_ACCOUNT','COP','ACTIVE','2026-10-27T00:00:00Z','2026-10-27T00:00:00Z')`); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO movements (id, financial_event_id, account_id, source_entry_id, movement_date, amount_minor, status, created_at, updated_at) VALUES ('mv-imp','fe-cents','acc-cents','se-1','2026-10-27',100000,'ACTIVE','2026-10-27T00:00:00Z','2026-10-27T00:00:00Z')`); err != nil {
		t.Fatalf("seed imported movement: %v", err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO movements (id, financial_event_id, account_id, movement_date, amount_minor, status, created_at, updated_at) VALUES ('mv-man','fe-cents','acc-cents','2026-10-27',7,'ACTIVE','2026-10-27T00:00:00Z','2026-10-27T00:00:00Z')`); err != nil {
		t.Fatalf("seed manual movement: %v", err)
	}

	// Apply again: this is what every startup does.
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("re-apply migrations: %v", err)
	}

	var imported int64
	if err := sqlDB.QueryRow("SELECT amount_minor FROM movements WHERE id = 'mv-imp'").Scan(&imported); err != nil {
		t.Fatalf("read imported movement: %v", err)
	}
	if imported != 10000000 {
		t.Errorf("an imported movement must be rescaled exactly once (100000 -> 10000000), got %d", imported)
	}
	var manual int64
	if err := sqlDB.QueryRow("SELECT amount_minor FROM movements WHERE id = 'mv-man'").Scan(&manual); err != nil {
		t.Fatalf("read manual movement: %v", err)
	}
	if manual != 7 {
		t.Errorf("a manual movement was already centavos and must not move, got %d", manual)
	}
}
