package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/log"
	"financeapp/backend/internal/repositories"
)

// reviewResolver is satisfied by app.ReviewService. Same rule as
// movementWriter: api cannot import app, so the transaction owner is injected.
type reviewResolver interface {
	ListEntries(ctx context.Context) ([]repositories.ReviewEntry, error)
	Resolve(ctx context.Context, entryID string, description string, amountMinor int64, date string, eventType string, accountID string, categoryID *string) (domain.Movement, error)
}

func (rt *router) listReviewEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := rt.reviews.ListEntries(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to load review queue"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, entries)
}

func (rt *router) resolveReviewEntry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Description string  `json:"description"`
		AmountMinor int64   `json:"amount_minor"`
		Date        string  `json:"date"`
		Type        string  `json:"type"`
		AccountID   string  `json:"account_id"`
		CategoryID  *string `json:"category_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if req.AccountID == "" || req.Date == "" || req.Description == "" {
		http.Error(w, `{"error":"account_id, date and description are required"}`, http.StatusBadRequest)
		return
	}
	// The event_type CHECK in the schema accepts INCOME and EXPENSE: validate
	// here so the caller gets a clear message instead of a constraint crash.
	if req.Type != "INCOME" && req.Type != "EXPENSE" {
		http.Error(w, `{"error":"type must be INCOME or EXPENSE"}`, http.StatusBadRequest)
		return
	}

	created, err := rt.reviews.Resolve(r.Context(), id, req.Description, req.AmountMinor, req.Date, req.Type, req.AccountID, req.CategoryID)
	if err != nil {
		log.Info(fmt.Sprintf("[REVIEW RESOLVE] failed id=%s err=%v", id, err))
		writeRepoError(w, err)
		return
	}
	log.Info(fmt.Sprintf("[REVIEW RESOLVE] entry=%s movement=%s", id, created.ID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encodeJSON(w, created)
}
