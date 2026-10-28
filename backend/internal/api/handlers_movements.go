package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/log"
)

// movementWriter is satisfied by app.MovementService. The api package cannot
// import app (app imports api), so the transaction owner arrives injected.
type movementWriter interface {
	CreateWithEvent(ctx context.Context, fe domain.FinancialEvent, mv domain.Movement, components []domain.MovementComponent) (domain.Movement, error)
	CreateManual(ctx context.Context, cmd domain.ManualMovementCommand) (domain.ManualMovementOutcome, error)
}

// transferWriter is satisfied by app.TransferService. Same injection rule:
// the transfer policy (two legs, one relation, counterpart-aware delete)
// lives in the service; the handler maps it to HTTP.
type transferWriter interface {
	Transfer(ctx context.Context, cmd domain.TransferCommand) (domain.TransferOutcome, error)
	DeleteMovement(ctx context.Context, movementID string) error
}

func (rt *router) listMovements(w http.ResponseWriter, r *http.Request) {
	movements, err := rt.movementRepo.ListWithCategory(r.Context(), listLimit(r))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list movements")
		return
	}
	type movementDTO struct {
		domain.Movement
		CategoryName           string `json:"categoryName"`
		CategoryID             string `json:"categoryId,omitempty"`
		CounterpartAccountName string `json:"counterpartAccountName,omitempty"`
		Currency               string `json:"currency"`
	}
	out := make([]movementDTO, 0, len(movements))
	for _, m := range movements {
		out = append(out, movementDTO{Movement: m.Movement, CategoryName: m.CategoryName, CategoryID: m.CategoryID, CounterpartAccountName: m.CounterpartAccountName, Currency: m.AccountCurrency})
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
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	log.Info(fmt.Sprintf("[MOVEMENT CREATE] account=%s amount=%d type=%s date=%s desc=%s", req.AccountID, req.AmountMinor, req.Type, req.Date, req.Description))
	// A zero amount is neither income nor expense: same criterion as the
	// parser, the manual path cannot fabricate financial identity either.
	if req.AmountMinor == 0 {
		writeJSONError(w, http.StatusBadRequest, "amount_minor must not be zero")
		return
	}
	// Validate the event type here so the caller gets a clear message: the
	// column has no CHECK, so anything else would be stored as-is and fail
	// much later, far from the cause.
	switch req.Type {
	case "INCOME", "EXPENSE", "TRANSFER":
	default:
		writeJSONError(w, http.StatusBadRequest, "type must be INCOME, EXPENSE or TRANSFER")
		return
	}

	// Duplicate detection, review parking and the atomic write are the
	// service's policy; this handler maps the outcome to HTTP.
	if req.Type == "TRANSFER" && req.ToAccountID != "" {
		// One decision with two legs: the transfer service owns it.
		outcome, err := rt.transfers.Transfer(r.Context(), domain.TransferCommand{
			OriginID:      req.AccountID,
			DestinationID: req.ToAccountID,
			AmountMinor:   req.AmountMinor,
			Date:          req.Date,
			Description:   req.Description,
		})
		if err != nil {
			if errors.Is(err, domain.ErrTransferToSelf) {
				writeJSONError(w, http.StatusBadRequest, "a transfer needs two different accounts")
				return
			}
			if errors.Is(err, domain.ErrTransferAmount) {
				writeJSONError(w, http.StatusBadRequest, "the transfer amount must be positive")
				return
			}
			log.Info(fmt.Sprintf("[MOVEMENT CREATE] transfer failed: %v", err))
			writeRepoError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		encodeJSON(w, map[string]string{"id": outcome.OriginMovementID, "counterpart_id": outcome.DestinationMovementID})
		return
	}

	outcome, err := rt.movements.CreateManual(r.Context(), domain.ManualMovementCommand{
		AccountID:   req.AccountID,
		AmountMinor: req.AmountMinor,
		Type:        req.Type,
		ToAccountID: req.ToAccountID,
		Description: req.Description,
		Date:        req.Date,
		CategoryID:  req.CategoryID,
		Force:       force,
	})
	if err != nil {
		log.Info(fmt.Sprintf("[MOVEMENT CREATE] transactional create failed: %v", err))
		writeRepoError(w, err)
		return
	}
	if outcome.Duplicate {
		writeJSONObject(w, http.StatusConflict, map[string]string{
			"error":          "duplicate",
			"redirect":       "/review",
			"review_item_id": outcome.ParkedReviewID,
		})
		return
	}
	log.Info(fmt.Sprintf("[MOVEMENT CREATE] Created movement %s", outcome.ID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encodeJSON(w, map[string]string{"id": outcome.ID})
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
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	mv, err := rt.movementRepo.GetByID(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "movement not found")
		return
	}
	if mv.FinancialEventID == "" {
		writeJSONError(w, http.StatusBadRequest, "no financial event")
		return
	}
	if err := rt.eventRepo.SetCategory(r.Context(), mv.FinancialEventID, body.CategoryID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rt *router) deleteMovement(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Transfer-aware: deleting either leg deletes the whole transfer.
	if err := rt.transfers.DeleteMovement(r.Context(), id); err != nil {
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
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if mv.AccountID != accountID {
		log.Info(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] mismatch account=%s movement_account=%s", accountID, mv.AccountID))
		writeJSONError(w, http.StatusForbidden, "unauthorized")
		return
	}
	// Transfer-aware, like the plain delete: a leg deletes its transfer.
	if err := rt.transfers.DeleteMovement(r.Context(), mid); err != nil {
		log.Info(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] error=%v", err))
		writeRepoError(w, err)
		return
	}
	log.Info(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] success id=%s", mid))
	w.WriteHeader(http.StatusNoContent)
}
