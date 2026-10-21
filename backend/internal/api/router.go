package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	_ "modernc.org/sqlite"
	"time"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
	logpkg "financeapp/backend/internal/log"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func NewRouter(cfg config.Config, docRepo *repositories.DocumentRepository, importRepo *repositories.ImportJobRepo, accountRepo *repositories.AccountRepository, movementRepo *repositories.MovementRepository, eventRepo *repositories.FinancialEventRepository, sourceEntryRepo *repositories.SourceEntryRepository, categoryRepo *repositories.CategoryRepository, confirmHandler func(context.Context, string) error) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w, map[string]string{"status":"ok"})
	})
	uploadHandler := NewUploadHandler(cfg, docRepo, importRepo, accountRepo, movementRepo, eventRepo, sourceEntryRepo)
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
		encodeJSON(w,inst)
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
		encodeJSON(w,created)
	})
	mux.HandleFunc("POST /jobs", CreateJobHandler)
	mux.HandleFunc("GET /jobs/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/jobs/"):]
		GetJobHandler(w, r, importRepo, id)
	})
	mux.HandleFunc("GET /jobs/{id}/parsed", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		// Load job to get processing_id
		job, err := importRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		if job.ProcessingID == nil {
			http.Error(w, `{"error":"processing not ready"}`, http.StatusBadRequest)
			return
		}
		artifactPath := filepath.Join(cfg.ArtifactsPath, *job.ProcessingID, "parsed.json")
		data, err := os.ReadFile(artifactPath)
		if err != nil {
			http.Error(w, `{"error":"parsed not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
	mux.HandleFunc("GET /jobs/{id}/manifest", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		job, err := importRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		if job.ProcessingID == nil {
			http.Error(w, `{"error":"processing not ready"}`, http.StatusBadRequest)
			return
		}
		artifactPath := filepath.Join(cfg.ArtifactsPath, *job.ProcessingID, "manifest.json")
		data, err := os.ReadFile(artifactPath)
		if err != nil {
			http.Error(w, `{"error":"manifest not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
	mux.HandleFunc("POST /jobs/{id}/confirm", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		job, err := importRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		if job.Status == "SUCCESS" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"SUCCESS","id":"`+id+`"}`))
			return
		}
		if job.Status == "REVIEW" {
			// check pending decisions - simplified: reject if status is REVIEW
			http.Error(w, `{"error":"decisions pending"}`, http.StatusConflict)
			return
		}
		// Call real import service via injected handler
		if confirmHandler != nil {
			if err := confirmHandler(r.Context(), id); err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
				return
			}
		}
		// reload job
		job, _ = importRepo.GetByID(r.Context(), id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"`+job.Status+`","id":"`+id+`"}`))
	})
	mux.HandleFunc("GET /accounts", func(w http.ResponseWriter, r *http.Request) {
		accounts, err := accountRepo.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list accounts"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,accounts)
	})
	mux.HandleFunc("GET /movements", func(w http.ResponseWriter, r *http.Request) {
		movements, err := movementRepo.List(r.Context())
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
			fe, _ := eventRepo.GetByID(r.Context(), m.FinancialEventID)
			catName := ""
			catID := ""
			if fe.CategoryID != nil {
				catID = *fe.CategoryID
				if cat, err := categoryRepo.GetByID(r.Context(), *fe.CategoryID); err == nil {
					catName = cat.Name
				}
			}
			out = append(out, movementDTO{Movement: m, CategoryName: catName, CategoryID: catID})
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,out)
	})
	mux.HandleFunc("POST /movements", func(w http.ResponseWriter, r *http.Request) {
		logpkg.LogMsg("[MOVEMENT CREATE] Request received")
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
			logpkg.LogMsg("[MOVEMENT CREATE] Invalid body: " + err.Error())
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}
		logpkg.LogMsg(fmt.Sprintf("[MOVEMENT CREATE] account=%s amount=%d type=%s date=%s desc=%s", req.AccountID, req.AmountMinor, req.Type, req.Date, req.Description))
		if !force {
			existing, _ := movementRepo.List(r.Context())
			for _, m := range existing {
				if m.AccountID == req.AccountID && m.MovementDate == req.Date && m.AmountMinor == req.AmountMinor && (m.Description == nil || req.Description == "" || *m.Description == req.Description) {
					logpkg.LogMsg(fmt.Sprintf("[MOVEMENT CREATE] Duplicate found %s", m.ID))
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusConflict)
					encodeJSON(w,map[string]string{"error":"duplicate","redirect":"/review"})
					return
				}
			}
		}
		now := time.Now().UTC().Format(time.RFC3339)
		feID := fmt.Sprintf("fe_manual_%d", time.Now().UnixNano())
		var catID *string
		if req.CategoryID != "" {
			catID = &req.CategoryID
		}
		fe := domain.FinancialEvent{
			ID: feID,
			EventType: req.Type,
			EventDate: &req.Date,
			Status: "ACTIVE",
			VerificationStatus: "UNVERIFIED",
			Description: &req.Description,
			CategoryID: catID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if _, err := eventRepo.Create(r.Context(), fe); err != nil {
			logpkg.LogMsg(fmt.Sprintf("[MOVEMENT CREATE] financial event create failed: %v", err))
			writeRepoError(w, err)
			return
		}
		mvID := fmt.Sprintf("mv_manual_%d", time.Now().UnixNano())
		mv := domain.Movement{
			ID: mvID,
			FinancialEventID: feID,
			AccountID: req.AccountID,
			MovementDate: req.Date,
			AmountMinor: req.AmountMinor,
			Description: &req.Description,
			Status: "ACTIVE",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if _, err := movementRepo.Create(r.Context(), mv, []domain.MovementComponent{}); err != nil {
			logpkg.LogMsg(fmt.Sprintf("[MOVEMENT CREATE] movement create failed: %v", err))
			writeRepoError(w, err)
			return
		}
		logpkg.LogMsg(fmt.Sprintf("[MOVEMENT CREATE] Created movement %s", mvID))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		encodeJSON(w,map[string]string{"id": mvID})
	})
	mux.HandleFunc("GET /documents", func(w http.ResponseWriter, r *http.Request) {
		documents, err := docRepo.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed to list documents"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,documents)
	})

	// Read-only traceability chain: Movement -> Financial Event -> Source Entry -> Document
	mux.HandleFunc("GET /documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		doc, err := docRepo.GetByID(r.Context(), r.PathValue("id"))
		if err != nil {
			writeLookupError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,doc)
	})
	mux.HandleFunc("GET /financial-events/{id}", func(w http.ResponseWriter, r *http.Request) {
		event, err := eventRepo.GetByID(r.Context(), r.PathValue("id"))
		if err != nil {
			writeLookupError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,event)
	})
	mux.HandleFunc("GET /source-entries/{id}", func(w http.ResponseWriter, r *http.Request) {
		entry, err := sourceEntryRepo.GetByID(r.Context(), r.PathValue("id"))
		if err != nil {
			writeLookupError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,entry)
	})

	// Account entity + its movements and documents (read-only)
	mux.HandleFunc("GET /accounts/{id}", func(w http.ResponseWriter, r *http.Request) {
		account, err := accountRepo.GetByID(r.Context(), r.PathValue("id"))
		if err != nil {
			writeLookupError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,account)
	})
	mux.HandleFunc("GET /accounts/{id}/movements", func(w http.ResponseWriter, r *http.Request) {
		movements, err := movementRepo.ListByAccount(r.Context(), r.PathValue("id"))
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
			fe, _ := eventRepo.GetByID(r.Context(), m.FinancialEventID)
			catName := ""
			catID := ""
			if fe.CategoryID != nil {
				catID = *fe.CategoryID
				if cat, err := categoryRepo.GetByID(r.Context(), *fe.CategoryID); err == nil {
					catName = cat.Name
				}
			}
			out = append(out, movementDTO{Movement: m, CategoryName: catName, CategoryID: catID})
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,out)
	})
	mux.HandleFunc("GET /accounts/{id}/documents", func(w http.ResponseWriter, r *http.Request) {
		documents, err := docRepo.ListByAccount(r.Context(), r.PathValue("id"))
		if err != nil {
			http.Error(w, `{"error":"failed to list documents"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,documents)
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
		encodeJSON(w,docs)
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
		type movementDTO struct {
			domain.Movement
			CategoryName string `json:"categoryName"`
		}
		out := make([]movementDTO, 0, len(movs))
		for _, m := range movs {
			fe, _ := eventRepo.GetByID(r.Context(), m.FinancialEventID)
			catName := ""
			if fe.CategoryID != nil {
				if cat, err := categoryRepo.GetByID(r.Context(), *fe.CategoryID); err == nil {
					catName = cat.Name
				}
			}
			out = append(out, movementDTO{Movement: m, CategoryName: catName})
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,out)
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
		encodeJSON(w,out)
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
		catNames := make(map[string]string)
		for _, m := range movements {
			if from != "" && m.MovementDate < from { continue }
			if to != "" && m.MovementDate > to { continue }
			if m.Status != "ACTIVE" { continue }
			if m.FinancialEventID == "" { continue }
			evt, err := eventRepo.GetByID(r.Context(), m.FinancialEventID)
			if err != nil { continue }
			if evt.EventType != "INCOME" { continue }
			var catName string
			if evt.CategoryID == nil || *evt.CategoryID == "" {
				catName = "Sin categoría"
			} else {
				catID := *evt.CategoryID
				cat, err := categoryRepo.GetByID(r.Context(), catID)
				if err != nil {
					catName = "Sin categoría"
				} else {
					catName = cat.Name
				}
			}
			key := catName
			dist[key] += m.AmountMinor
			catNames[key] = catName
		}
		var out []map[string]interface{}
		for k, v := range dist {
			out = append(out, map[string]interface{}{"category": k, "amount_minor": v})
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,out)
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
			var catName string
			if evt.CategoryID == nil || *evt.CategoryID == "" {
				catName = "Sin categoría"
			} else {
				catID := *evt.CategoryID
				cat, err := categoryRepo.GetByID(r.Context(), catID)
				if err != nil {
					catName = "Sin categoría"
				} else {
					catName = cat.Name
				}
			}
			key := catName
			dist[key] += -m.AmountMinor
		}
		var out []map[string]interface{}
		for k, v := range dist {
			out = append(out, map[string]interface{}{"category": k, "amount_minor": v})
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,out)
	})

	// Admin purge endpoint
	mux.HandleFunc("DELETE /admin/purge", func(w http.ResponseWriter, r *http.Request) {
		// Simple dev purge - deletes all domain data using Go driver
		dbPath := cfg.DatabasePath
		logDir := cfg.LogsPath
		logErr := func(msg string) {
			logFile := filepath.Join(logDir, time.Now().Format("02-01-2006.log"))
			if f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); f != nil {
				fmt.Fprintf(f, "[%s] PURGE %s: %s\n", time.Now().Format("15:04:05"), msg, dbPath)
				f.Close()
			}
		}
		sqlDB, err := sql.Open("sqlite", dbPath)
		if err != nil {
			logErr("open failed " + err.Error())
			http.Error(w, fmt.Sprintf(`{"error":"purge failed","detail":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		defer sqlDB.Close()
		if _, err := sqlDB.Exec("PRAGMA foreign_keys=OFF"); err != nil {
			logErr("pragma failed " + err.Error())
			http.Error(w, fmt.Sprintf(`{"error":"purge failed","detail":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		tables := []string{
			"source_entries",
			"movements",
			"financial_events",
			"documents",
			"import_jobs",
			"accounts",
			"categories",
		}
		for _, t := range tables {
			if _, err := sqlDB.Exec("DELETE FROM " + t); err != nil {
				// Log error to file
				logFile := filepath.Join(logDir, time.Now().Format("02-01-2006.log"))
				if f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); f != nil {
					fmt.Fprintf(f, "[%s] PURGE error table %s: %v\n", time.Now().Format("15:04:05"), t, err)
					f.Close()
				}
				http.Error(w, fmt.Sprintf(`{"error":"purge failed","detail":"table %s: %v"}`, t, err), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})

	// Categories CRUD
	mux.HandleFunc("GET /categories", func(w http.ResponseWriter, r *http.Request) {
		cats, err := categoryRepo.List(r.Context(), "ACTIVE")
		if err != nil {
			http.Error(w, `{"error":"failed to list categories"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,cats)
	})
	mux.HandleFunc("POST /categories", func(w http.ResponseWriter, r *http.Request) {
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
		created, err := categoryRepo.Create(r.Context(), cat)
		if err != nil {
			http.Error(w, `{"error":"failed to create category"}`, http.StatusInternalServerError)
			return
		}
		logpkg.LogMsg(fmt.Sprintf("[CATEGORY CREATE] %s %s", created.ID, created.Name))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		encodeJSON(w,created)
	})
	mux.HandleFunc("PUT /categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var req struct {
			Name         string `json:"name"`
			CategoryType string `json:"category_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}
		existing, err := categoryRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		// Log review trigger
		logpkg.LogMsg(fmt.Sprintf("[CATEGORY UPDATE] %s -> %s", id, req.Name))
		logpkg.LogMsg(fmt.Sprintf("[REVIEW TRIGGER] Category %s edited, movements with this category should be reviewed", id))
		w.Header().Set("Content-Type", "application/json")
		encodeJSON(w,existing)
	})
	mux.HandleFunc("DELETE /categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		// Soft delete for safety, but we can physically delete for now
		dbPath := os.Getenv("DATABASE_PATH")
		if dbPath == "" {
			if appdata := os.Getenv("APPDATA"); appdata != "" {
				dbPath = filepath.Join(appdata, "FinanceApp", "data", "db", "finance.db")
			} else {
				dbPath = "./data/finance.db"
			}
		}
		dbConn, err := sql.Open("sqlite", dbPath)
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
		logpkg.LogMsg(fmt.Sprintf("[CATEGORY DELETE] %s", id))
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("PUT /movements/{id}/category", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body struct { CategoryID *string `json:"categoryId"` }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}
		mv, err := movementRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"movement not found"}`, http.StatusNotFound)
			return
		}
		if mv.FinancialEventID == "" {
			http.Error(w, `{"error":"no financial event"}`, http.StatusBadRequest)
			return
		}
		dbPath := os.Getenv("DATABASE_PATH")
		if dbPath == "" {
			if appdata := os.Getenv("APPDATA"); appdata != "" {
				dbPath = filepath.Join(appdata, "FinanceApp", "data", "db", "finance.db")
			} else {
				dbPath = "./data/finance.db"
			}
		}
		dbConn, err := sql.Open("sqlite", dbPath)
		if err != nil { http.Error(w, `{"error":"db open"}`, http.StatusInternalServerError); return }
		defer dbConn.Close()
		_, err = dbConn.ExecContext(r.Context(), "UPDATE financial_events SET category_id = ?, updated_at = ? WHERE id = ?", body.CategoryID, time.Now().UTC().Format(time.RFC3339), mv.FinancialEventID)
		if err != nil { http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError); return }
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("DELETE /movements/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		dbPath := os.Getenv("DATABASE_PATH")
		if dbPath == "" {
			if appdata := os.Getenv("APPDATA"); appdata != "" {
				dbPath = filepath.Join(appdata, "FinanceApp", "data", "db", "finance.db")
			} else {
				dbPath = "./data/finance.db"
			}
		}
		dbConn, err := sql.Open("sqlite", dbPath)
		if err != nil { http.Error(w, `{"error":"db"}`, http.StatusInternalServerError); return }
		defer dbConn.Close()
		_, _ = dbConn.ExecContext(r.Context(), "DELETE FROM movements WHERE id = ?", id)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("DELETE /accounts/{id}/movements/{mid}", func(w http.ResponseWriter, r *http.Request) {
		accountID := r.PathValue("id")
		mid := r.PathValue("mid")
		logpkg.LogMsg(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] account=%s movement=%s", accountID, mid))
		// verify movement belongs to account
		mv, err := movementRepo.GetByID(r.Context(), mid)
		if err != nil {
			logpkg.LogMsg(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] not found id=%s err=%v", mid, err))
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		if mv.AccountID != accountID {
			logpkg.LogMsg(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] mismatch account=%s movement_account=%s", accountID, mv.AccountID))
			http.Error(w, `{"error":"unauthorized"}`, http.StatusForbidden)
			return
		}
		if err := movementRepo.Delete(r.Context(), mid); err != nil {
			logpkg.LogMsg(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] error=%v", err))
			http.Error(w, `{"error":"failed to delete"}`, http.StatusInternalServerError)
			return
		}
		logpkg.LogMsg(fmt.Sprintf("[MOVEMENT DELETE ACCOUNT] success id=%s", mid))
		w.WriteHeader(http.StatusNoContent)
	})
	// Logging middleware
	loggedMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logpkg.Info(fmt.Sprintf("HTTP %s %s start", r.Method, r.URL.Path))
		rw := &responseWriter{w, http.StatusOK}
		mux.ServeHTTP(rw, r)
		duration := time.Since(start)
		logpkg.Info(fmt.Sprintf("HTTP %s %s %d %s", r.Method, r.URL.Path, rw.status, duration))
	})
	return corsMiddleware(loggedMux)
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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
