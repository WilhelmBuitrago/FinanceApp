package api

import (
	"crypto/sha256"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/log"
	"financeapp/backend/internal/repositories"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type UploadResponse struct {
	DocumentID   string `json:"document_id"`
	JobID        string `json:"job_id"`
	ProcessingID string `json:"processing_id,omitempty"`
}

type UploadHandler struct {
	cfg         config.Config
	docRepo     *repositories.DocumentRepository
	jobRepo     *repositories.ImportJobRepo
	accountRepo *repositories.AccountRepository
	run         ProcessorRunnerFunc
	imports     ImportRunner
}

func NewUploadHandler(cfg config.Config, docRepo *repositories.DocumentRepository, jobRepo *repositories.ImportJobRepo, accountRepo *repositories.AccountRepository, run ProcessorRunnerFunc, imports ImportRunner) *UploadHandler {
	return &UploadHandler{cfg: cfg, docRepo: docRepo, jobRepo: jobRepo, accountRepo: accountRepo, run: run, imports: imports}
}

func (h *UploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Info("UploadHandler Handle start")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Error(fmt.Sprintf("ParseMultipartForm error: %v", err))
		http.Error(w, "invalid multipart", http.StatusBadRequest)
		return
	}
	accountID := r.FormValue("account_id")
	if accountID == "" {
		log.Error("account_id missing")
		http.Error(w, `{"error":"account_id required"}`, http.StatusBadRequest)
		return
	}
	entity := r.FormValue("entity")
	// Validate account exists and active
	acc, err := h.accountRepo.GetByID(r.Context(), accountID)
	if err != nil {
		log.Error(fmt.Sprintf("account not found %s: %v", accountID, err))
		http.Error(w, `{"error":"account not found"}`, http.StatusBadRequest)
		return
	}
	if acc.Status != "ACTIVE" {
		log.Error(fmt.Sprintf("account not active %s status=%s", accountID, acc.Status))
		http.Error(w, `{"error":"account not active"}`, http.StatusBadRequest)
		return
	}
	if acc.Institution == nil || *acc.Institution != "NU" {
		log.Error(fmt.Sprintf("institution not supported %s", accountID))
		http.Error(w, `{"error":"institution not supported"}`, http.StatusBadRequest)
		return
	}
	// Optional entity verification for accessibility
	if entity != "" && acc.Institution != nil && entity != *acc.Institution {
		log.Error(fmt.Sprintf("entity mismatch account %s entity=%s institution=%s", accountID, entity, *acc.Institution))
		http.Error(w, `{"error":"entity does not match account institution"}`, http.StatusBadRequest)
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		log.Error(fmt.Sprintf("file missing: %v", err))
		http.Error(w, "file missing", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	docsBase := h.cfg.DocumentsPath
	dstPath := filepath.Join(docsBase, "raw", hdr.Filename)
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		http.Error(w, "cannot create dir", http.StatusInternalServerError)
		return
	}
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "cannot create file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = dst.Close() }()

	hasher := sha256.New()
	mw := io.TeeReader(file, hasher)
	if _, err := io.Copy(dst, mw); err != nil {
		http.Error(w, "save failed", http.StatusInternalServerError)
		return
	}
	sha := fmt.Sprintf("%x", hasher.Sum(nil))

	now := time.Now().UTC().Format(time.RFC3339)
	docID := "doc_" + sha[:8]
	inst := *acc.Institution
	document := domain.Document{
		ID:           docID,
		AccountID:    &accountID,
		DocumentType: "PDF",
		Institution:  &inst,
		FileHash:     &sha,
		SourceUri:    &dstPath,
		Status:       "UPLOADED",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	// The same PDF twice is a conflict, not a crash: docID is derived from
	// the file hash, so re-uploading an imported file must not 500.
	if _, err := h.docRepo.GetByID(r.Context(), docID); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		encodeJSON(w, map[string]string{"error": "document already imported", "document_id": docID})
		return
	}
	if _, err := h.docRepo.Create(r.Context(), document); err != nil {
		http.Error(w, fmt.Sprintf("document persist failed: %v", err), http.StatusInternalServerError)
		return
	}

	jobID := "job_" + sha[:8]
	job := domain.ImportJob{
		ID:         jobID,
		DocumentID: docID,
		Status:     "UPLOAD",
		CreatedAt:  now,
	}
	if _, err := h.jobRepo.Create(r.Context(), job); err != nil {
		http.Error(w, "job persist failed", http.StatusInternalServerError)
		return
	}

	started := time.Now().UTC().Format(time.RFC3339)
	job.Status = "PROCESSING"
	job.StartedAt = &started
	if _, err := h.jobRepo.Update(r.Context(), job); err != nil {
		http.Error(w, "job update failed", http.StatusInternalServerError)
		return
	}

	run := h.run
	if run == nil {
		run = RunProcessor
	}
	res, err := run(dstPath, accountID)
	if err != nil {
		finished := time.Now().UTC().Format(time.RFC3339)
		job.Status = "ERROR"
		job.FinishedAt = &finished
		e := err.Error()
		job.Error = &e
		if _, err := h.jobRepo.Update(r.Context(), job); err != nil {
			log.Error(fmt.Sprintf("failed to mark job ERROR: %v", err))
		}
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf(`{"error":"processor failed","detail":%q}`, e), http.StatusInternalServerError)
		return
	}

	job.ProcessingID = &res.ProcessingID
	if _, err := h.jobRepo.Update(r.Context(), job); err != nil {
		http.Error(w, "job update failed", http.StatusInternalServerError)
		return
	}

	// Update document status to PROCESSED
	doc, err := h.docRepo.GetByID(r.Context(), docID)
	if err == nil {
		doc.Status = "PROCESSED"
		doc.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		if _, err := h.docRepo.Update(r.Context(), doc); err != nil {
			log.Error(fmt.Sprintf("document status update to PROCESSED failed: %v", err))
		}
	}

	// Staging is owned by the import service: one transaction, review items
	// included. This handler only validates, persists and reports.
	if h.imports == nil {
		http.Error(w, "import service not configured", http.StatusInternalServerError)
		return
	}
	if err := h.imports.Import(r.Context(), jobID); err != nil {
		finished := time.Now().UTC().Format(time.RFC3339)
		job.Status = "FAILED"
		job.FinishedAt = &finished
		e := err.Error()
		job.Error = &e
		if _, updErr := h.jobRepo.Update(r.Context(), job); updErr != nil {
			log.Error(fmt.Sprintf("failed to record job failure: %v", updErr))
		}
		http.Error(w, fmt.Sprintf(`{"error":"import failed","detail":%q}`, e), http.StatusInternalServerError)
		return
	}

	// Update document status to IMPORTED
	docUpd, err := h.docRepo.GetByID(r.Context(), docID)
	if err == nil {
		docUpd.Status = "IMPORTED"
		docUpd.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		if _, err := h.docRepo.Update(r.Context(), docUpd); err != nil {
			log.Error(fmt.Sprintf("document status update to IMPORTED failed: %v", err))
		}
	}

	resp := UploadResponse{
		DocumentID:   docID,
		JobID:        jobID,
		ProcessingID: res.ProcessingID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encodeJSON(w, resp)
}

func GetJobHandler(w http.ResponseWriter, r *http.Request, repo *repositories.ImportJobRepo, id string) {
	job, err := repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, job)
}
