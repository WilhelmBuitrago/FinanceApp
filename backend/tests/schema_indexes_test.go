package tests

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"financeapp/backend/internal/db"
)

/* The schema contract the listings depend on: every column the hot
   queries filter and sort by carries an index. A listing ordered by
   created_at over a table without that index is a full scan wearing a
   costume; these names are the promise the queries rely on. */

var expectedIndexes = map[string]string{
	"idx_movements_account_id":         "movements",
	"idx_movements_financial_event_id": "movements",
	"idx_movements_movement_date":      "movements",
	"idx_movements_created_at":         "movements",
}

func TestMigrations_CreateTheIndexesTheListingsNeed(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "indexes.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	for name, table := range expectedIndexes {
		var got string
		err := sqlDB.QueryRow("SELECT tbl_name FROM sqlite_master WHERE type = 'index' AND name = ?", name).Scan(&got)
		if err != nil {
			t.Fatalf("index %s on %s is missing: %v", name, table, err)
		}
		if got != table {
			t.Errorf("index %s must live on %s, got %s", name, table, got)
		}
	}

	// An index that exists but the planner ignores is a rumor: the hot
	// listing must read through the created_at index, not scan the table.
	var detail string
	var id, parent, notUsed int
	if err := sqlDB.QueryRow(
		"EXPLAIN QUERY PLAN SELECT id FROM movements ORDER BY created_at DESC",
	).Scan(&id, &parent, &notUsed, &detail); err != nil {
		t.Fatalf("explain the listing: %v", err)
	}
	if !strings.Contains(detail, "USING INDEX") {
		t.Errorf("the created_at DESC listing must read through its index, plan: %s", detail)
	}
}
