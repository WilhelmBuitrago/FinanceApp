package repositories

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// Enable foreign keys per connection
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign_keys: %v", err)
	}
	if err := applyMigrations(t, sqlDB); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	return sqlDB
}

// applyMigrations applies every migration in internal/db/migrations in
// filename order. The directory is resolved from this file's location, not
// from the process working directory, so tests can run from any package.
func applyMigrations(t *testing.T, sqlDB *sql.DB) error {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("cannot resolve test helper location")
	}
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "db", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if err != nil {
			return err
		}
		stmts := strings.Split(string(content), ";")
		for _, s := range stmts {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if _, err := sqlDB.Exec(s); err != nil {
				return fmt.Errorf("%s: %w", entry.Name(), err)
			}
		}
	}
	return nil
}

func NewTestPersistence(t *testing.T) *Persistence {
	t.Helper()
	sqlDB := setupTestDB(t)
	p := NewPersistence(sqlDB)
	return p
}
