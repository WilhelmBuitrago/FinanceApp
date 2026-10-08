package repositories

import (
	"financeapp/backend/internal/db"
)

type DB struct {
	Querier db.Querier
}

func NewSQLiteRepository(dbPath string) (*DB, error) {
	// Stub initialization. In production, open sqlite connection here.
	// Example with modernc.org/sqlite:
	// sqlDB, err := sql.Open("sqlite", dbPath)
	// if err != nil { return nil, err }
	// ...
	// For now return empty DB to satisfy compilation.
	return &DB{
		Querier: nil,
	}, nil
}
