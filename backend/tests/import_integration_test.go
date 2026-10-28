package tests

import (
	"database/sql"
	"financeapp/backend/internal/db"
	"testing"

	_ "modernc.org/sqlite"
)

func TestImportIntegrationValid(t *testing.T) {
	dbPath := "file::memory:?cache=shared"
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sqlDB.Close() }()
	_, _ = sqlDB.Exec("PRAGMA foreign_keys=ON")
	// apply migrations
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// basic sanity
	var cnt int
	_ = sqlDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='import_jobs'").Scan(&cnt)
	if cnt != 1 {
		t.Fatalf("import_jobs missing")
	}
}
