package tests

import (
	"context"
	"financeapp/backend/internal/repositories"
	"testing"
)

func TestReadModelList(t *testing.T) {
	// This is a smoke test – in production we'd spin up an in-memory SQLite.
	// Here we only verify that List methods compile and return a slice without panicking
	// when Querier is nil – they should return repositories.ErrUnavailable.
	ctx := context.Background()
	repo := repositories.NewAccountRepository(&repositories.DB{Querier: nil})
	_, err := repo.List(ctx)
	if err == nil {
		t.Fatalf("expected repositories.ErrUnavailable")
	}
}
