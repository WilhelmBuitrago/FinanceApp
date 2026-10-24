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
	ListPending(ctx context.Context) ([]repositories.ReviewItemRecord, error)
	Submit(ctx context.Context, kind string, payload any, sourceEntryID *string) (string, error)
	Kind(ctx context.Context, itemID string) (string, error)
	Resolve(ctx context.Context, itemID string, description string, amountMinor int64, date string, eventType string, accountID string, categoryID *string) (domain.Movement, error)
	Discard(ctx context.Context, itemID string) error
	ResolvePeriod(ctx context.Context, itemID string, month, year int) error
}

func (rt *router) listReviewEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := rt.reviews.ListPending(r.Context())
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
	// The correction shape depends on the kind: a CATEGORY item only moves
	// the category on an existing movement, so demanding the movement
	// fields here made it unresolvable (the frontend correctly sends only
	// category_id). The kinds that CREATE a movement keep requiring the
	// full correction.
	kind, err := rt.reviews.Kind(r.Context(), id)
	if err != nil {
		writeLookupError(w, err)
		return
	}
	if kind != "CATEGORY" {
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

// discardReviewEntry parks an item as discarded: the human decided it is not
// real (nothing is written to the financial data).
func (rt *router) discardReviewEntry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := rt.reviews.Discard(r.Context(), id); err != nil {
		log.Info(fmt.Sprintf("[REVIEW DISCARD] failed id=%s err=%v", id, err))
		writeRepoError(w, err)
		return
	}
	log.Info(fmt.Sprintf("[REVIEW DISCARD] entry=%s", id))
	w.WriteHeader(http.StatusNoContent)
}

// resolveReviewPeriod closes a DOCUMENT_PERIOD item: the human supplies the
// statement month and year (ABR 2026), and the backend derives the full
// period range and the parked movement dates of that document against it.
func (rt *router) resolveReviewPeriod(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Month int `json:"month"`
		Year  int `json:"year"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if req.Month < 1 || req.Month > 12 || req.Year < 2000 || req.Year > 2100 {
		http.Error(w, `{"error":"month 1-12 and a four-digit year are required"}`, http.StatusBadRequest)
		return
	}
	if err := rt.reviews.ResolvePeriod(r.Context(), id, req.Month, req.Year); err != nil {
		log.Info(fmt.Sprintf("[REVIEW PERIOD] failed id=%s err=%v", id, err))
		writeRepoError(w, err)
		return
	}
	log.Info(fmt.Sprintf("[REVIEW PERIOD] resolved id=%s month=%d year=%d", id, req.Month, req.Year))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"RESOLVED","id":"` + id + `"}`))
}
