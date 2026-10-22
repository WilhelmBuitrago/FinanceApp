package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"crypto/sha256"
	"time"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/domain/entities"
	"financeapp/backend/internal/repositories"
	"financeapp/backend/internal/log"
)

type UploadResponse struct {
	DocumentID   string `json:"document_id"`
	JobID        string `json:"job_id"`
	ProcessingID string `json:"processing_id,omitempty"`
}

type UploadHandler struct {
	cfg            config.Config
	docRepo         *repositories.DocumentRepository
	jobRepo         *repositories.ImportJobRepo
	accountRepo     *repositories.AccountRepository
	movementRepo    *repositories.MovementRepository
	eventRepo       *repositories.FinancialEventRepository
	sourceEntryRepo *repositories.SourceEntryRepository
}

func NewUploadHandler(cfg config.Config, docRepo *repositories.DocumentRepository, jobRepo *repositories.ImportJobRepo, accountRepo *repositories.AccountRepository, movementRepo *repositories.MovementRepository, eventRepo *repositories.FinancialEventRepository, sourceEntryRepo *repositories.SourceEntryRepository) *UploadHandler {
	return &UploadHandler{cfg: cfg, docRepo: docRepo, jobRepo: jobRepo, accountRepo: accountRepo, movementRepo: movementRepo, eventRepo: eventRepo, sourceEntryRepo: sourceEntryRepo}
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
	defer file.Close()

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
	defer dst.Close()

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
	if _, err := h.docRepo.Create(r.Context(), document); err != nil {
		http.Error(w, fmt.Sprintf("document persist failed: %v", err), http.StatusInternalServerError)
		return
	}

	jobID := "job_" + sha[:8]
	job := entities.ImportJob{
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

	res, err := RunProcessor(dstPath, accountID)
	if err != nil {
		finished := time.Now().UTC().Format(time.RFC3339)
		job.Status = "ERROR"
		job.FinishedAt = &finished
		e := err.Error()
		job.Error = &e
		h.jobRepo.Update(r.Context(), job)
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

	// Stage source entries from parsed artifact
	artifactsBase := h.cfg.ArtifactsPath
	artifactPath := filepath.Join(artifactsBase, res.ProcessingID, "parsed.json")
	if data, err := os.ReadFile(artifactPath); err == nil {
		var parsed struct {
			Movements []struct {
				DateOriginal      string `json:"date_original"`
				DateNormalized    string `json:"date_normalized"`
				DescriptionOriginal string `json:"description_original"`
				AmountNormalized  int    `json:"amount_normalized"`
				PageNumber        int    `json:"page_number"`
				Sequence          int    `json:"sequence"`
			} `json:"movements"`
		}
		if json.Unmarshal(data, &parsed) == nil && h.sourceEntryRepo != nil {
			now := time.Now().UTC().Format(time.RFC3339)
			for _, m := range parsed.Movements {
				seID := fmt.Sprintf("%s_se_%d_%d", docID, m.PageNumber, m.Sequence)
				se := domain.SourceEntry{
					ID:                    seID,
					DocumentID:            docID,
					RawDescription:        &m.DescriptionOriginal,
					RawDate:               &m.DateOriginal,
					NormalizedDescription: &m.DescriptionOriginal,
					NormalizedDate:        &m.DateNormalized,
					PageNumber:            int64Ptr(int64(m.PageNumber)),
					RowNumber:             int64Ptr(int64(m.Sequence)),
					ExtractionID:          &res.ProcessingID,
					CreatedAt:             now,
					UpdatedAt:             now,
				}
				if _, err := h.sourceEntryRepo.Create(r.Context(), se); err != nil {
				h.failJob(w, r, job, fmt.Sprintf("staging source entry failed: %v", err))
				return
			}

				// Auto-import: create financial event and movement
				eventType := "INCOME"
				if m.AmountNormalized < 0 {
					eventType = "EXPENSE"
				}
				feID := fmt.Sprintf("%s_fe_%d_%d", docID, m.PageNumber, m.Sequence)
				fe := domain.FinancialEvent{
					ID:                 feID,
					EventType:         eventType,
					EventDate:         sPtr(m.DateNormalized),
					Status:            "ACTIVE",
					VerificationStatus: "UNVERIFIED",
					Description:       sPtr(m.DescriptionOriginal),
					CreatedAt:         now,
					UpdatedAt:         now,
				}
				if _, err := h.eventRepo.Create(r.Context(), fe); err != nil {
				h.failJob(w, r, job, fmt.Sprintf("staging financial event failed: %v", err))
				return
			}

				mvID := fmt.Sprintf("%s_mv_%d_%d", docID, m.PageNumber, m.Sequence)
				mv := domain.Movement{
					ID:               mvID,
					FinancialEventID: feID,
					AccountID:        accountID,
					SourceEntryID:    &seID,
					MovementDate:     m.DateNormalized,
					AmountMinor:      int64(m.AmountNormalized),
					Description:      sPtr(m.DescriptionOriginal),
					Status:           "ACTIVE",
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				if _, err := h.movementRepo.Create(r.Context(), mv); err != nil {
				h.failJob(w, r, job, fmt.Sprintf("staging movement failed: %v", err))
				return
			}
			}
		}
	}

	// Job moves to SUCCESS automatically
	job.Status = "SUCCESS"
	finished := time.Now().UTC().Format(time.RFC3339)
	job.FinishedAt = &finished
	if _, err := h.jobRepo.Update(r.Context(), job); err != nil {
		http.Error(w, "job update failed", http.StatusInternalServerError)
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

func CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(`{"message":"job creation scaffold"}`)); err != nil {
		log.Error(fmt.Sprintf("job scaffold response write failed: %v", err))
	}
}

func GetJobHandler(w http.ResponseWriter, r *http.Request, repo *repositories.ImportJobRepo, id string) {
	job, err := repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w,job)
}

func int64Ptr(v int64) *int64 { return &v }
func sPtr(s string) *string { return &s }

// failJob marks the import job as FAILED and reports the failure to the
// caller: a partially staged import must never look like a success.
func (h *UploadHandler) failJob(w http.ResponseWriter, r *http.Request, job entities.ImportJob, reason string) {
	job.Status = "FAILED"
	job.Error = &reason
	finished := time.Now().UTC().Format(time.RFC3339)
	job.FinishedAt = &finished
	if _, err := h.jobRepo.Update(r.Context(), job); err != nil {
		log.Error(fmt.Sprintf("failed to record job failure: %v", err))
	}
	http.Error(w, fmt.Sprintf(`{"error":%q}`, reason), http.StatusInternalServerError)
}


