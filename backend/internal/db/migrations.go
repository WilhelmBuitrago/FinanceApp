package db

import (
	"database/sql"
	"embed"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Apply runs every migration embedded in migrations/ in filename order.
// The directory is read on each call, so a new migration file is applied
// without touching this list (the test helper resolves the same set).
// Each file is split on ';' and executed statement by statement. Errors
// meaning "already applied" (already exists, duplicate column) are ignored
// so re-running Apply on an existing database is a no-op.
func Apply(sqlDB *sql.DB) error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		for _, s := range strings.Split(string(content), ";") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if _, err := sqlDB.Exec(s); err != nil {
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
