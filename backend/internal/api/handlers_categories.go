package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"database/sql"

	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/log"
	_ "modernc.org/sqlite"
)

func (rt *router) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := rt.categoryRepo.List(r.Context(), "ACTIVE")
	if err != nil {
		http.Error(w, `{"error":"failed to list categories"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, cats)
}

func (rt *router) createCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		CategoryType string `json:"category_type"`
		ParentID     *string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
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
		http.Error(w, `{"error":"failed to create category"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	existing, err := rt.categoryRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
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
	// Soft delete for safety, but we can physically delete for now
	dbConn, err := sql.Open("sqlite", rt.cfg.DatabasePath)
	if err != nil {
		http.Error(w, `{"error":"db open failed"}`, http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()
	_, err = dbConn.ExecContext(r.Context(), "DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error":"delete failed"}`, http.StatusInternalServerError)
		return
	}
	log.Info(fmt.Sprintf("[CATEGORY DELETE] %s", id))
	w.WriteHeader(http.StatusNoContent)
}
