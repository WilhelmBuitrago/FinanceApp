package db

import (
	"database/sql"
	"os"
	"strings"
)

func Apply(sqlDB *sql.DB) error {
	migrations := []string{
		"db/migrations/0001_initial_schema.sql",
		"db/migrations/0002_import_jobs.sql",
		"db/migrations/0003_import_review_staging.sql",
	}
	for _, path := range migrations {
		content, err := os.ReadFile(path)
		if err != nil {
			alt := "../..//" + path
			content, err = os.ReadFile(alt)
			if err != nil {
				return err
			}
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
