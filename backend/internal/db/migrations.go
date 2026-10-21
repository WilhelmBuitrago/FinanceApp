package db

import (
	"database/sql"
	"embed"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Apply(sqlDB *sql.DB) error {
	migrations := []string{
		"migrations/0001_initial_schema.sql",
		"migrations/0002_import_jobs.sql",
		"migrations/0003_import_review_staging.sql",
		"migrations/0004_align_schema_with_generated_queries.sql",
	}
	for _, path := range migrations {
		content, err := migrationsFS.ReadFile(path)
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
				// Idempotent: ignore "already exists" and duplicate column
				errMsg := err.Error()
				if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate column") {
					continue
				}
				return err
			}
		}
	}
	return nil
}
