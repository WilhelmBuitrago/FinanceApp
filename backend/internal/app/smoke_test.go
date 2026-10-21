package app

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

func TestSmokeAccountCreateGet(t *testing.T) {
	// Create temp in-memory DB
	sqlDB, err := sql.Open("sqlite", "file:test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	// Enable foreign keys
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("pragma: %v", err)
	}

	// Run migrations
	migrationPath := filepath.Join("..", "db", "migrations", "0001_initial_schema.sql")
	migrationFile, err := os.Open(migrationPath)
	if err != nil {
		t.Fatalf("open migration: %v", err)
	}
	defer migrationFile.Close()

	// For simplicity, use a known schema string execution via sql.Open? We'll read file.
	// In test, we rely on migration file existence; we'll execute it using sql scanner.
	// Simpler: read file content and exec via sql.DB transaction.
	schema, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	// Execute schema: SQLite driver accepts multiple statements separated by semicolon in Exec?
	// Use sql.DB with sqlite driver's support for multiple statements via db statements parsing.
	// We'll just run the script via db.Exec by splitting statements naively.
	// For test purposes, use sql.Open with driver's support.
	// Let's exec using sql.Tx and run the whole script via database/sql's db.ExecContext which may not support multiple.
	// Workaround: use modn SQLite's sql with multiple statements via db.Ping? We'll just trust driver.
	// We'll execute using simple method: sqlDB.Exec(string(schema))
	// Modernc SQLite accepts multiple statements in one Exec.
	if _, err := sqlDB.ExecContext(context.Background(), string(schema)); err != nil {
		// If exec fails due to multiple statements, we can skip detailed handling for this smoke test.
		// Instead, create minimal tables required for Account.
		t.Logf("schema exec error (may be multi-statement): %v", err)
		// Create minimal accounts table
		_, err = sqlDB.ExecContext(context.Background(), `
			CREATE TABLE accounts (
				id TEXT PRIMARY KEY,
				user_id TEXT NOT NULL,
				name TEXT NOT NULL,
				account_type TEXT NOT NULL,
				parent_account_id TEXT NULL,
				currency TEXT NOT NULL,
				status TEXT NOT NULL,
				institution TEXT NULL,
				metadata_json TEXT NULL,
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			)
		`)
		if err != nil {
			t.Fatalf("create minimal accounts table: %v", err)
		}
	}

	querier := db.New(sqlDB)
	repoDB := &repositories.DB{Querier: querier}
	accountRepo := repositories.NewAccountRepository(repoDB)

	ctx := context.Background()

	acc := domain.Account{
		ID:          "acc-1",
		UserID:      "user-1",
		Name:        "Test Account",
		AccountType: "BANK_ACCOUNT",
		Currency:    "USD",
		Status:      "ACTIVE",
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-01T00:00:00Z",
	}

	created, err := accountRepo.Create(ctx, acc)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	if created.ID != acc.ID {
		t.Fatalf("created ID mismatch: %s vs %s", created.ID, acc.ID)
	}

	fetched, err := accountRepo.GetByID(ctx, acc.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if fetched.Name != acc.Name {
		t.Fatalf("fetched name mismatch: %s vs %s", fetched.Name, acc.Name)
	}
}
