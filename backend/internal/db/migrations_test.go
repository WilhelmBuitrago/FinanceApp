package db

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestApplyMigrations(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer sqlDB.Close()
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("pragma: %v", err)
	}
	if err := Apply(sqlDB); err != nil {
		t.Fatalf("apply: %v", err)
	}
	// idempotent
	if err := Apply(sqlDB); err != nil {
		t.Fatalf("apply twice: %v", err)
	}
	tables := []string{"documents", "financial_events", "movements", "import_jobs"}
	for _, tbl := range tables {
		var name string
		err := sqlDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tbl).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", tbl, err)
		}
	}
	// FK check
	var fk sql.NullString
	err = sqlDB.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='import_jobs'`).Scan(&fk)
	if err != nil {
		t.Fatalf("query import_jobs schema: %v", err)
	}
	if !strings.Contains(fk.String, "REFERENCES documents(id)") {
		t.Fatalf("import_jobs FK missing")
	}
}
