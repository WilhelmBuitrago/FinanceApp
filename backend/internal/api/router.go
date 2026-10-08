package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"time"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

func NewRouter(docRepo *repositories.DocumentRepository, importRepo *repositories.ImportJobRepo, accountRepo *repositories.AccountRepository, movementRepo *repositories.MovementRepository, eventRepo *repositories.FinancialEventRepository, sourceEntryRepo *repositories.SourceEntryRepository) http.Handler {
	mux := http.NewServeMux()
	uploadHandler := NewUploadHandler(docRepo, importRepo, accountRepo, movementRepo, eventRepo)
	mux.HandleFunc("POST /documents/upload", uploadHandler.Handle)
	mux.HandleFunc("GET /institutions", func(w http.ResponseWriter, r *http.Request) {
		inst := []map[string]interface{}{
			{
				"code": "NU",
				"name": "Nu",
				"supported_account_types": []string{"BANK_ACCOUNT", "CREDIT_CARD"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(inst)
	})
	mux.HandleFunc("POST /accounts", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID              string  `json:"id"`
			UserID          string  `json:"user_id"`
			Name            string  `json:"name"`
			AccountType     string  `json:"account_type"`
			Currency        string  `json:"currency"`
			Status          string  `json:"status"`
			Institution     string  `json:"institution"`
			MetadataJson    string  `json:"metadata_json,omitempty"`
			ValidationData  string  `json:"validation_data,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}
		// Basic validation per Sprint 12
		if req.Institution != "NU" {
			http.Error(w, `{"error":"institution not supported"}`, http.StatusBadRequest)
			return
		}
		if req.AccountType != "BANK_ACCOUNT" && req.AccountType != "CREDIT_CARD" {
			http.Error(w, `{"error":"account type not supported"}`, http.StatusBadRequest)
			return
		}
		if req.Status == "" {
			req.Status = "ACTIVE"
		}
		if req.Currency == "" {
			req.Currency = "COP"
		}
		if req.ID == "" {
			http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
			return
		}
		now := timeNowUTC()
		meta := req.MetadataJson
		if req.ValidationData != "" {
			// store validation data masked in metadata for now
			meta = fmt.Sprintf(`{"validation_data":"%s"}`, req.ValidationData)
		}
		acc := domain.Account{
			ID:          req.ID,
			UserID:      req.UserID,
			Name:        req.Name,
			AccountType: req.AccountType,
			Currency:    req.Currency,
			Status:      req.Status,
			Institution: &req.Institution,
			MetadataJson: &meta,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		created, err := accountRepo.Create(r.Context(), acc)
		if err != nil {
			http.Error(w, `{"error":"failed to create account"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
	})
	mux.HandleFunc("POST /jobs", CreateJobHandler)
	mux.HandleFunc("GET /jobs/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/jobs/"):]
		GetJobHandler(w, r, importRepo, id)
	})
	mux.HandleFunc("GET /accounts", func(w http.ResponseWriter, r *http.Request) {
		accounts, err := accountRepo.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list accounts"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(accounts)
	})
	mux.HandleFunc("GET /movements", func(w http.ResponseWriter, r *http.Request) {
		movements, err := movementRepo.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(movements)
	})
	mux.HandleFunc("GET /documents", func(w http.ResponseWriter, r *http.Request) {
		documents, err := docRepo.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list documents"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(documents)
	})

	// Read-only traceability chain: Movement -> Financial Event -> Source Entry -> Document
	mux.HandleFunc("GET /documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		doc, err := docRepo.GetByID(r.Context(), r.PathValue("id"))
		if err != nil {
			writeLookupError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(doc)
	})
	mux.HandleFunc("GET /financial-events/{id}", func(w http.ResponseWriter, r *http.Request) {
		event, err := eventRepo.GetByID(r.Context(), r.PathValue("id"))
		if err != nil {
			writeLookupError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(event)
	})
	mux.HandleFunc("GET /source-entries/{id}", func(w http.ResponseWriter, r *http.Request) {
		entry, err := sourceEntryRepo.GetByID(r.Context(), r.PathValue("id"))
		if err != nil {
			writeLookupError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entry)
	})

	// Account entity + its movements and documents (read-only)
	mux.HandleFunc("GET /accounts/{id}", func(w http.ResponseWriter, r *http.Request) {
		account, err := accountRepo.GetByID(r.Context(), r.PathValue("id"))
		if err != nil {
			writeLookupError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(account)
	})
	mux.HandleFunc("GET /accounts/{id}/movements", func(w http.ResponseWriter, r *http.Request) {
		movements, err := movementRepo.ListByAccount(r.Context(), r.PathValue("id"))
		if err != nil {
			http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(movements)
	})
	mux.HandleFunc("GET /accounts/{id}/documents", func(w http.ResponseWriter, r *http.Request) {
		documents, err := docRepo.ListByAccount(r.Context(), r.PathValue("id"))
		if err != nil {
			http.Error(w, `{"error":"failed to list documents"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(documents)
	})

	mux.HandleFunc("GET /documents/recent", func(w http.ResponseWriter, r *http.Request) {
		docs, err := docRepo.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list documents"}`, http.StatusInternalServerError)
			return
		}
		// limit 3
		if len(docs) > 3 {
			docs = docs[:3]
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(docs)
	})

	mux.HandleFunc("GET /movements/recent", func(w http.ResponseWriter, r *http.Request) {
		movs, err := movementRepo.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
			return
		}
		if len(movs) > 5 {
			movs = movs[:5]
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(movs)
	})

	// Dashboard aggregates
	mux.HandleFunc("GET /dashboard/cash-flow", func(w http.ResponseWriter, r *http.Request) {
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		movements, err := movementRepo.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
			return
		}
		// Simple in-memory aggregation
		agg := make(map[string]struct{inflow, outflow int64})
		for _, m := range movements {
			if from != "" && m.MovementDate < from {
				continue
			}
			if to != "" && m.MovementDate > to {
				continue
			}
			// Skip cancelled
			if m.Status != "ACTIVE" {
				continue
			}
			date := m.MovementDate
			cur := agg[date]
			if m.AmountMinor > 0 {
				cur.inflow += m.AmountMinor
			} else {
				cur.outflow += -m.AmountMinor
			}
			agg[date] = cur
		}
		type point struct {
			Date   string `json:"date"`
			Inflow int64  `json:"inflow"`
			Outflow int64 `json:"outflow"`
			Net    int64  `json:"net"`
		}
		var out []point
		for d, v := range agg {
			out = append(out, point{Date: d, Inflow: v.inflow, Outflow: v.outflow, Net: v.inflow - v.outflow})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	mux.HandleFunc("GET /dashboard/income-distribution", func(w http.ResponseWriter, r *http.Request) {
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		movements, err := movementRepo.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
			return
		}
		dist := make(map[string]int64)
		for _, m := range movements {
			if from != "" && m.MovementDate < from { continue }
			if to != "" && m.MovementDate > to { continue }
			if m.Status != "ACTIVE" { continue }
			if m.FinancialEventID == "" { continue }
			evt, err := eventRepo.GetByID(r.Context(), m.FinancialEventID)
			if err != nil { continue }
			if evt.EventType != "INCOME" { continue }
			key := "uncategorized"
			if evt.CategoryID != nil && *evt.CategoryID != "" {
				key = *evt.CategoryID
			}
			dist[key] += m.AmountMinor
		}
		var out []map[string]interface{}
		for k, v := range dist {
			out = append(out, map[string]interface{}{"category": k, "amount_minor": v})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	mux.HandleFunc("GET /dashboard/expense-distribution", func(w http.ResponseWriter, r *http.Request) {
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		movements, err := movementRepo.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list movements"}`, http.StatusInternalServerError)
			return
		}
		dist := make(map[string]int64)
		for _, m := range movements {
			if from != "" && m.MovementDate < from { continue }
			if to != "" && m.MovementDate > to { continue }
			if m.Status != "ACTIVE" { continue }
			if m.FinancialEventID == "" { continue }
			evt, err := eventRepo.GetByID(r.Context(), m.FinancialEventID)
			if err != nil { continue }
			if evt.EventType != "EXPENSE" && evt.EventType != "DEBIT_PURCHASE" && evt.EventType != "CREDIT_PAYMENT" {
				continue
			}
			key := "uncategorized"
			if evt.CategoryID != nil && *evt.CategoryID != "" {
				key = *evt.CategoryID
			}
			dist[key] += -m.AmountMinor // make positive
		}
		var out []map[string]interface{}
		for k, v := range dist {
			out = append(out, map[string]interface{}{"category": k, "amount_minor": v})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	// Admin purge endpoint
	mux.HandleFunc("DELETE /admin/purge", func(w http.ResponseWriter, r *http.Request) {
		// Simple dev purge - deletes all domain data.
		// DB path is CWD-relative, matching main.go's DATABASE_PATH default
		// (the server runs from backend/).
		cmd := exec.Command("C:\\platform-tools\\sqlite3.exe", "data\\finance.db",
			"DELETE FROM movement_components; DELETE FROM event_relations; DELETE FROM movement_relations; DELETE FROM installments; DELETE FROM installment_plans; DELETE FROM source_entries; DELETE FROM movements; DELETE FROM financial_events; DELETE FROM documents; DELETE FROM import_jobs; DELETE FROM accounts; DELETE FROM categories;")
		if err := cmd.Run(); err != nil {
			http.Error(w, `{"error":"purge failed"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})

	return corsMiddleware(mux)
}

func writeLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, repositories.ErrNotFound) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	http.Error(w, `{"error":"lookup failed"}`, http.StatusInternalServerError)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Allow Vite dev server and Tauri desktop origins
		allowedOrigins := map[string]bool{
			"http://localhost:5173": true,
			"http://localhost:5174": true,
			"http://127.0.0.1:5173": true,
			"http://127.0.0.1:5174": true,
		}
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func contextWithJobID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, "job_id", id)
}

func timeNowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}
