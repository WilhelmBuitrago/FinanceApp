package tests

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/repositories"
)

// newTestPersistence wires a migrated in-memory database for the black-box
// tests. It applies the SAME embedded migrations the binary applies
// (db.Apply): the old helper reimplemented the applier statement by
// statement, which is exactly the kind of duplication that drifts.
func newTestPersistence(t *testing.T) *repositories.Persistence {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// Enable foreign keys per connection.
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign_keys: %v", err)
	}
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return repositories.NewPersistence(sqlDB)
}

// Shared pointer helpers for the tests in this folder.
func strPtr(s string) *string { return &s }

func int64Ptr(v int64) *int64 { return &v }
