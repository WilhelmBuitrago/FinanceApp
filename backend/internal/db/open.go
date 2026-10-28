package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
)

/* OpenSQLite opens the application database with the pragmas the
   workload needs. Two writes racing the same SQLite file is not a
   corner case here: the desktop posts a movement and a document import
   can be writing at the same time, and without a busy timeout the
   loser gets "database is locked" as an HTTP 500. WAL lets a reader
   proceed while a writer commits, instead of failing the read.

   Both the server and the admin purge open the database; both go
   through here so the timeout is a property of the file, not of the
   caller that remembered to set it. */

func OpenSQLite(path string) (*sql.DB, error) {
	dsn := "file:" + filepath.ToSlash(path) + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// One writer serializes in WAL; a pool bigger than that only makes
	// connections queue on the timeout.
	sqlDB.SetMaxOpenConns(4)
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return sqlDB, nil
}
