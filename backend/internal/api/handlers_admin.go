package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"financeapp/backend/internal/db"
)

// adminPurge deletes every domain table directly on the SQLite database the
// configuration points to (dev/maintenance endpoint: no repository cascade
// runs here on purpose).
func (rt *router) adminPurge(w http.ResponseWriter, r *http.Request) {
	// Simple dev purge - deletes all domain data using Go driver
	dbPath := rt.cfg.DatabasePath
	logDir := rt.cfg.LogsPath
	logErr := func(msg string) {
		logFile := filepath.Join(logDir, time.Now().Format("02-01-2006.log"))
		if f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); f != nil {
			_, _ = fmt.Fprintf(f, "[%s] PURGE %s: %s\n", time.Now().Format("15:04:05"), msg, dbPath)
			_ = f.Close()
		}
	}
	sqlDB, err := db.OpenSQLite(dbPath)
	if err != nil {
		logErr("open failed " + err.Error())
		writeJSONObject(w, http.StatusInternalServerError, map[string]string{"error": "purge failed", "detail": err.Error()})
		return
	}
	defer func() { _ = sqlDB.Close() }()
	if _, err := sqlDB.Exec("PRAGMA foreign_keys=OFF"); err != nil {
		logErr("pragma failed " + err.Error())
		writeJSONObject(w, http.StatusInternalServerError, map[string]string{"error": "purge failed", "detail": err.Error()})
		return
	}
	tables := []string{
		"source_entries",
		"movements",
		"financial_events",
		"documents",
		"import_jobs",
		"accounts",
		"categories",
	}
	for _, t := range tables {
		if _, err := sqlDB.Exec("DELETE FROM " + t); err != nil {
			// Log error to file
			logFile := filepath.Join(logDir, time.Now().Format("02-01-2006.log"))
			if f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); f != nil {
				_, _ = fmt.Fprintf(f, "[%s] PURGE error table %s: %v\n", time.Now().Format("15:04:05"), t, err)
				_ = f.Close()
			}
			writeJSONObject(w, http.StatusInternalServerError, map[string]string{"error": "purge failed", "detail": fmt.Sprintf("table %s: %v", t, err)})
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
