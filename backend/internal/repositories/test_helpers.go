package repositories

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"os"
	"strings"
	"testing"

	"financeapp/backend/internal/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := ":memory:"
	sqlDB, err := sql.Open("sqlite", dbPath)
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

func applyMigrations(t *testing.T, sqlDB *sql.DB) error {
	t.Helper()
	migrations := []string{
		"../db/migrations/0001_initial_schema.sql",
		"../db/migrations/0002_import_jobs.sql",
		"../db/migrations/0003_import_review_staging.sql",
		"../db/migrations/0004_align_schema_with_generated_queries.sql",
		"../db/migrations/0005_expand_component_types.sql",
	}
	for _, path := range migrations {
		content, err := os.ReadFile(path)
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
				return err
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

func NewTestQuerier(t *testing.T) db.Querier {
	t.Helper()
	sqlDB := setupTestDB(t)
	return db.New(sqlDB)
}

func NewTestDBQuerierPersistence(t *testing.T) (*sql.DB, db.Querier, *Persistence) {
	t.Helper()
	sqlDB := setupTestDB(t)
	querier := db.New(sqlDB)
	persistence := NewPersistence(sqlDB)
	return sqlDB, querier, persistence
}
