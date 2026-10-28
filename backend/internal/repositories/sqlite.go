package repositories

import (
	"financeapp/backend/internal/db"
)

// DB bundles the sqlc querier shared by every repository.
type DB struct {
	Querier db.Querier
}
