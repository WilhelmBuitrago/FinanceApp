package tests

import (
	"database/sql"
	"path/filepath"
	"testing"
)

/* Every test that writes opens its database the way the server opens
   the real one: busy timeout and WAL. The transfer burst surfaced
   SQLITE_BUSY against a bare DSN — the same 500 the desktop would
   have seen, here as a failure nobody could explain. This helper makes
   the pragmas the test environment's default too, so the contract
   under test is the production one. */

func openTestSQLite(t *testing.T, name string) *sql.DB {
	t.Helper()
	dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), name)) + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlDB
}
