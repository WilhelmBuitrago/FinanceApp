package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/log"
)

// movementWriter is satisfied by app.MovementService. The api package cannot
// import app (app imports api), so the transaction owner arrives injected.
type movementWriter interface {
	CreateWithEvent(ctx context.Context, fe domain.FinancialEvent, mv domain.Movement, components []domain.MovementComponent) (domain.Movement, error)
}

func (rt *router) listMovements(w http.ResponseWriter, r *http.Request) {
	movements, err := rt.movementRepo.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
		return
	}
	type movementDTO struct {
		domain.Movement
		CategoryName string `json:"categoryName"`
		CategoryID   string `json:"categoryId,omitempty"`
	}
	out := make([]movementDTO, 0, len(movements))
	for _, m := range movements {
		fe, _ := rt.eventRepo.GetByID(r.Context(), m.FinancialEventID)
		catName := ""
		catID := ""
		if fe.CategoryID != nil {
			catID = *fe.CategoryID
			if cat, err := rt.categoryRepo.GetByID(r.Context(), *fe.CategoryID); err == nil {
				catName = cat.Name
			}
		}
		out = append(out, movementDTO{Movement: m, CategoryName: catName, CategoryID: catID})
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, out)
}

func (rt *router) createMovement(w http.ResponseWriter, r *http.Request) {
	log.Info("[MOVEMENT CREATE] Request received")
	force := r.URL.Query().Get("force") == "true"
	var req struct {
		AmountMinor int64  `json:"amount_minor"`
		Type        string `json:"type"`
		AccountID   string `json:"account_id"`
		ToAccountID string `json:"to_account_id,omitempty"`
		Description string `json:"description"`
		Date        string `json:"date"`
		CategoryID  string `json:"category_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Info("[MOVEMENT CREATE] Invalid body: " + err.Error())
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	log.Info(fmt.Sprintf("[MOVEMENT CREATE] account=%s amount=%d type=%s date=%s desc=%s", req.AccountID, req.AmountMinor, req.Type, req.Date, req.Description))
	// A zero amount is neither income nor expense: same criterion as the
	// parser, the manual path cannot fabricate financial identity either.
	if req.AmountMinor == 0 {
		http.Error(w, `{"error":"amount_minor must not be zero"}`, http.StatusBadRequest)
		return
	}
	// Validate the event type here so the caller gets a clear message: the
	// column has no CHECK, so anything else would be stored as-is and fail
	// much later, far from the cause.
	switch req.Type {
	case "INCOME", "EXPENSE", "TRANSFER":
	default:
		http.Error(w, `{"error":"type must be INCOME, EXPENSE or TRANSFER"}`, http.StatusBadRequest)
		return
	}
	if !force {
		existing, err := rt.movementRepo.List(r.Context())
		if err != nil {
			log.Info(fmt.Sprintf("[MOVEMENT CREATE] duplicate check failed: %v", err))
			http.Error(w, `{"error":"failed to check duplicates"}`, http.StatusInternalServerError)
			return
		}
		for _, m := range existing {
			if m.AccountID == req.AccountID && m.MovementDate == req.Date && m.AmountMinor == req.AmountMinor && (m.Description == nil || req.Description == "" || *m.Description == req.Description) {
				log.Info(fmt.Sprintf("[MOVEMENT CREATE] Duplicate found %s", m.ID))
				// Park the duplicate in the unified review queue so the human
				// decides (persisted, no localStorage).
				itemID, parkErr := rt.reviews.Submit(r.Context(), "DUPLICATE", map[string]any{
					"account_id":    req.AccountID,
					"amount_minor":  req.AmountMinor,
					"date":          req.Date,
					"type":          req.Type,
					"description":   req.Description,
					"collides_with": m.ID,
				}, nil)
				if parkErr != nil {
					log.Info(fmt.Sprintf("[MOVEMENT CREATE] failed to park duplicate: %v", parkErr))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				encodeJSON(w, map[string]string{
					"error":          "duplicate",
					"redirect":       "/review",
					"review_item_id": itemID,
				})
				return
			}
		}
	}
	now := timeNowUTC()
	feID := fmt.Sprintf("fe_manual_%d", time.Now().UnixNano())
	var catID *string
	if req.CategoryID != "" {
		catID = &req.CategoryID
	}
	fe := domain.FinancialEvent{
		ID:                 feID,
		EventType:          req.Type,
		EventDate:          &req.Date,
		Status:             "ACTIVE",
		VerificationStatus: "UNVERIFIED",
		Description:        &req.Description,
		CategoryID:         catID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	mvID := fmt.Sprintf("mv_manual_%d", time.Now().UnixNano())
	mv := domain.Movement{
		ID:               mvID,
		FinancialEventID: feID,
		AccountID:        req.AccountID,
		MovementDate:     req.Date,
		AmountMinor:      req.AmountMinor,
		Description:      &req.Description,
		Status:           "ACTIVE",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if _, err := rt.movements.CreateWithEvent(r.Context(), fe, mv, []domain.MovementComponent{}); err != nil {
		log.Info(fmt.Sprintf("[MOVEMENT CREATE] transactional create failed: %v", err))
		writeRepoError(w, err)
		return
	}
	log.Info(fmt.Sprintf("[MOVEMENT CREATE] Created movement %s", mvID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encodeJSON(w, map[string]string{"id": mvID})
}

// setMovementCategory moves the category to the financial event behind the
// movement (raw SQL on purpose: the category lives on the event, not the
// movement).
func (rt *router) setMovementCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		CategoryID *string `json:"categoryId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	mv, err := rt.movementRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"movement not found"}`, http.StatusNotFound)
		return
	}
	if mv.FinancialEventID == "" {
		http.Error(w, `{"error":"no financial event"}`, http.StatusBadRequest)
		return
	}
	if err := rt.eventRepo.SetCategory(r.Context(), mv.FinancialEventID, body.CategoryID); err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rt *router) deleteMovement(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := rt.movementRepo.Delete(r.Context(), id); err != nil {
		log.Info(fmt.Sprintf("[MOVEMENT DELETE] failed id=%s err=%v", id, err))
		writeRepoError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rt *router) deleteAccountMovement(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	mid := r.PathValue("mid")
	log.Info(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] account=%s movement=%s", accountID, mid))
	// verify movement belongs to account
	mv, err := rt.movementRepo.GetByID(r.Context(), mid)
	if err != nil {
		log.Info(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] not found id=%s err=%v", mid, err))
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if mv.AccountID != accountID {
		log.Info(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] mismatch account=%s movement_account=%s", accountID, mv.AccountID))
		http.Error(w, `{"error":"unauthorized"}`, http.StatusForbidden)
		return
	}
	if err := rt.movementRepo.Delete(r.Context(), mid); err != nil {
		log.Info(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] error=%v", err))
		http.Error(w, `{"error":"failed to delete"}`, http.StatusInternalServerError)
		return
	}
	log.Info(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] success id=%s", mid))
	w.WriteHeader(http.StatusNoContent)
}
