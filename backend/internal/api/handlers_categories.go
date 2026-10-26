package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/log"
)

func (rt *router) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := rt.categoryRepo.List(r.Context(), "ACTIVE")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, cats)
}

func (rt *router) createCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID           string  `json:"id"`
		Name         string  `json:"name"`
		CategoryType string  `json:"category_type"`
		ParentID     *string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	now := timeNowUTC()
	cat := domain.Category{
		ID:           req.ID,
		Name:         req.Name,
		CategoryType: &req.CategoryType,
		ParentID:     req.ParentID,
		Status:       "ACTIVE",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	created, err := rt.categoryRepo.Create(r.Context(), cat)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create category")
		return
	}
	log.Info(fmt.Sprintf("[CATEGORY CREATE] %s %s", created.ID, created.Name))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encodeJSON(w, created)
}

func (rt *router) updateCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name         string `json:"name"`
		CategoryType string `json:"category_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	existing, err := rt.categoryRepo.GetByID(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	// Log review trigger
	log.Info(fmt.Sprintf("[CATEGORY UPDATE] %s -> %s", id, req.Name))
	log.Info(fmt.Sprintf("[REVIEW TRIGGER] Category %s edited, movements with this category should be reviewed", id))
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, existing)
}

func (rt *router) deleteCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Capture the category and its movements first: the review payload needs
	// the name and the movements it leaves behind.
	category, catErr := rt.categoryRepo.GetByID(r.Context(), id)
	if catErr != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	affected, listErr := rt.movementRepo.ListByCategory(r.Context(), id)
	if listErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list affected movements")
		return
	}
	if err := rt.categoryRepo.Delete(r.Context(), id); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	log.Info(fmt.Sprintf("[CATEGORY DELETE] %s", id))
	// Park the movements left without a category in the unified review queue:
	// persisted and owned by the backend, no client-side bookkeeping.
	for _, m := range affected {
		if _, err := rt.reviews.Submit(r.Context(), "CATEGORY", map[string]any{
			"movement_id":   m.ID,
			"category_id":   id,
			"category_name": category.Name,
			"account_id":    m.AccountID,
			"amount_minor":  m.AmountMinor,
			"movement_date": m.MovementDate,
		}, nil); err != nil {
			log.Info(fmt.Sprintf("[CATEGORY DELETE] failed to park movement %s: %v", m.ID, err))
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
