package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"financeapp/backend/internal/domain"
)

func (rt *router) institutions(w http.ResponseWriter, r *http.Request) {
	inst := []map[string]interface{}{
		{
			"code":                    "NU",
			"name":                    "Nu",
			"supported_account_types": []string{"BANK_ACCOUNT", "CREDIT_CARD"},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, inst)
}

func (rt *router) createAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID             string `json:"id"`
		UserID         string `json:"user_id"`
		Name           string `json:"name"`
		AccountType    string `json:"account_type"`
		Currency       string `json:"currency"`
		Status         string `json:"status"`
		Institution    string `json:"institution"`
		MetadataJson   string `json:"metadata_json,omitempty"`
		ValidationData string `json:"validation_data,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	// Basic validation per Sprint 12
	if req.Institution != "NU" {
		writeJSONError(w, http.StatusBadRequest, "institution not supported")
		return
	}
	if req.AccountType != "BANK_ACCOUNT" && req.AccountType != "CREDIT_CARD" {
		writeJSONError(w, http.StatusBadRequest, "account type not supported")
		return
	}
	if req.Status == "" {
		req.Status = "ACTIVE"
	}
	if req.Currency == "" {
		req.Currency = "COP"
	}
	if req.ID == "" {
		writeJSONError(w, http.StatusBadRequest, "id required")
		return
	}
	now := timeNowUTC()
	meta := req.MetadataJson
	if req.ValidationData != "" {
		// store validation data masked in metadata for now
		meta = fmt.Sprintf(`{"validation_data":"%s"}`, req.ValidationData)
	}
	acc := domain.Account{
		ID:           req.ID,
		UserID:       req.UserID,
		Name:         req.Name,
		AccountType:  req.AccountType,
		Currency:     req.Currency,
		Status:       req.Status,
		Institution:  &req.Institution,
		MetadataJson: &meta,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	created, err := rt.accountRepo.Create(r.Context(), acc)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encodeJSON(w, created)
}

func (rt *router) listAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := rt.accountRepo.List(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, accounts)
}

// Account entity + its movements and documents (read-only)
func (rt *router) getAccount(w http.ResponseWriter, r *http.Request) {
	account, err := rt.accountRepo.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeLookupError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, account)
}

func (rt *router) listAccountMovements(w http.ResponseWriter, r *http.Request) {
	movements, err := rt.movementRepo.ListByAccount(r.Context(), r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list movements")
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

func (rt *router) listAccountDocuments(w http.ResponseWriter, r *http.Request) {
	documents, err := rt.docRepo.ListByAccount(r.Context(), r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list documents")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, documents)
}

// timeNowUTC is the shared RFC3339 UTC timestamp helper for handlers.
func timeNowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}
