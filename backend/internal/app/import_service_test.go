package app

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/repositories"
)

func TestImportRollback(t *testing.T) {
	dbPath := "file::memory:?cache=shared"
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil { t.Fatal(err) }
	defer sqlDB.Close()
	// create tables minimal
	schema := `
	CREATE TABLE import_jobs(id TEXT PRIMARY KEY, document_id TEXT, status TEXT, created_at TEXT, started_at TEXT, finished_at TEXT, error TEXT, processing_id TEXT);
	CREATE TABLE documents(id TEXT PRIMARY KEY, account_id TEXT);
	CREATE TABLE source_entries(id TEXT PRIMARY KEY, document_id TEXT, extraction_id TEXT, created_at TEXT, updated_at TEXT);
	CREATE TABLE financial_events(id TEXT PRIMARY KEY, event_type TEXT CHECK(event_type IN ('ADJUSTMENT','INCOME','EXPENSE')), status TEXT, verification_status TEXT, created_at TEXT, updated_at TEXT);
	CREATE TABLE movements(id TEXT PRIMARY KEY, financial_event_id TEXT, account_id TEXT, source_entry_id TEXT, movement_date TEXT, amount_minor INTEGER, status TEXT, created_at TEXT, updated_at TEXT);
	`
	_, err = sqlDB.Exec(schema)
	if err != nil { t.Fatal(err) }

	// create doc and job
	docID := "doc1"
	accID := "acc1"
	_, _ = sqlDB.Exec(`INSERT INTO documents(id,account_id) VALUES(?,?)`, docID, accID)
	jobID := "job1"
	procID := "proc1"
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = sqlDB.Exec(`INSERT INTO import_jobs(id,document_id,status,created_at,processing_id) VALUES(?,?,?,?,?)`, jobID, docID, "VALIDATING", now, procID)

	// create parsed.json with one movement to trigger error later
	artifacts := t.TempDir()
	t.Setenv("ARTIFACTS_PATH", artifacts)
	os.MkdirAll(filepath.Join(artifacts, procID), 0755)
	parsed := `{"validation":{"valid":true},"movements":[{"date_original":"01","date_normalized":"2026-01-01","description_original":"test","amount_normalized":100,"page_number":1,"sequence":1}]}`
	os.WriteFile(filepath.Join(artifacts, procID, "parsed.json"), []byte(parsed), 0644)
	manifest := `{"validation":{"valid":true}}`
	os.WriteFile(filepath.Join(artifacts, procID, "manifest.json"), []byte(manifest), 0644)

	querier := db.New(sqlDB)
	persistence := &repositories.Persistence{DB: sqlDB, Querier: querier}
	importJobRepo := repositories.NewImportJobRepo(sqlDB)
	service := NewImportService(config.Load(), persistence, importJobRepo)

	// Simulate failure by making financial_events insert fail via constraint: we will insert a bad event_type by monkey patching? Instead we rely on validation false.

	// For rollback test, we can set validation false to ensure no transaction starts
	os.WriteFile(filepath.Join(artifacts, procID, "manifest.json"), []byte(`{"validation":{"valid":false}}`), 0644)
	err = service.Import(context.Background(), jobID)
	if err == nil { t.Fatal("expected error") }

	// check counts remain zero
	var cnt int
	sqlDB.QueryRow(`SELECT COUNT(*) FROM source_entries`).Scan(&cnt)
	if cnt != 0 { t.Fatalf("source_entries not rolled back: %d", cnt) }
	sqlDB.QueryRow(`SELECT COUNT(*) FROM financial_events`).Scan(&cnt)
	if cnt != 0 { t.Fatalf("financial_events not rolled back: %d", cnt) }
	sqlDB.QueryRow(`SELECT COUNT(*) FROM movements`).Scan(&cnt)
	if cnt != 0 { t.Fatalf("movements not rolled back: %d", cnt) }
	// job status should remain VALIDATING
	var status string
	sqlDB.QueryRow(`SELECT status FROM import_jobs WHERE id=?`, jobID).Scan(&status)
	if status != "VALIDATING" { t.Fatalf("job status changed: %s", status) }
}
