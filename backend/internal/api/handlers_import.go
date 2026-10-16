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
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/domain/entities"
	"financeapp/backend/internal/repositories"
)

type UploadResponse struct {
	DocumentID   string `json:"document_id"`
	JobID        string `json:"job_id"`
	ProcessingID string `json:"processing_id,omitempty"`
}

type UploadHandler struct {
	docRepo         *repositories.DocumentRepository
	jobRepo         *repositories.ImportJobRepo
	accountRepo     *repositories.AccountRepository
	movementRepo    *repositories.MovementRepository
	eventRepo       *repositories.FinancialEventRepository
	sourceEntryRepo *repositories.SourceEntryRepository
}

func NewUploadHandler(docRepo *repositories.DocumentRepository, jobRepo *repositories.ImportJobRepo, accountRepo *repositories.AccountRepository, movementRepo *repositories.MovementRepository, eventRepo *repositories.FinancialEventRepository, sourceEntryRepo *repositories.SourceEntryRepository) *UploadHandler {
	return &UploadHandler{docRepo: docRepo, jobRepo: jobRepo, accountRepo: accountRepo, movementRepo: movementRepo, eventRepo: eventRepo, sourceEntryRepo: sourceEntryRepo}
}

func (h *UploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "invalid multipart", http.StatusBadRequest)
		return
	}
	accountID := r.FormValue("account_id")
	if accountID == "" {
		http.Error(w, `{"error":"account_id required"}`, http.StatusBadRequest)
		return
	}
	entity := r.FormValue("entity")
	// Validate account exists and active
	acc, err := h.accountRepo.GetByID(r.Context(), accountID)
	if err != nil {
		http.Error(w, `{"error":"account not found"}`, http.StatusBadRequest)
		return
	}
	if acc.Status != "ACTIVE" {
		http.Error(w, `{"error":"account not active"}`, http.StatusBadRequest)
		return
	}
	if acc.Institution == nil || *acc.Institution != "NU" {
		http.Error(w, `{"error":"institution not supported"}`, http.StatusBadRequest)
		return
	}
	// Optional entity verification for accessibility
	if entity != "" && acc.Institution != nil && entity != *acc.Institution {
		http.Error(w, `{"error":"entity does not match account institution"}`, http.StatusBadRequest)
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file missing", http.StatusBadRequest)
		return
	}
	defer file.Close()

	docsBase := os.Getenv("DOCUMENTS_PATH")
	if docsBase == "" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			docsBase = filepath.Join(appdata, "FinanceApp", "data", "documents")
		} else {
			docsBase = "data/documents"
		}
	}
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
		_, _ = h.docRepo.Update(r.Context(), doc)
	}

	// Stage source entries from parsed artifact
	artifactsBase := os.Getenv("ARTIFACTS_PATH")
	if artifactsBase == "" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			artifactsBase = filepath.Join(appdata, "FinanceApp", "data", "artifacts")
		} else {
			artifactsBase = "data/artifacts"
		}
	}
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
				_, _ = h.sourceEntryRepo.Create(r.Context(), se)

				// Auto-import: create financial event and movement
				eventType := "INCOME"
				if m.AmountNormalized < 0 {
					eventType = "EXPENSE"
				}
				feID := fmt.Sprintf("%s_fe_%d_%d", docID, m.PageNumber, m.Sequence)
				fe := domain.FinancialEvent{
					ID:                 feID,
					EventType:         eventType,
					EventDate:         strPtr(m.DateNormalized),
					Status:            "ACTIVE",
					VerificationStatus: "UNVERIFIED",
					Description:       strPtr(m.DescriptionOriginal),
					CreatedAt:         now,
					UpdatedAt:         now,
				}
				_, _ = h.eventRepo.Create(r.Context(), fe)

				mvID := fmt.Sprintf("%s_mv_%d_%d", docID, m.PageNumber, m.Sequence)
				mv := domain.Movement{
					ID:               mvID,
					FinancialEventID: feID,
					AccountID:        accountID,
					SourceEntryID:    &seID,
					MovementDate:     m.DateNormalized,
					AmountMinor:      int64(m.AmountNormalized),
					Description:      strPtr(m.DescriptionOriginal),
					Status:           "ACTIVE",
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				_, _ = h.movementRepo.Create(r.Context(), mv, []domain.MovementComponent{})
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
		_, _ = h.docRepo.Update(r.Context(), docUpd)
	}

	resp := UploadResponse{
		DocumentID:   docID,
		JobID:        jobID,
		ProcessingID: res.ProcessingID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`{"message":"job creation scaffold"}`))
}

func GetJobHandler(w http.ResponseWriter, r *http.Request, repo *repositories.ImportJobRepo, id string) {
	job, err := repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(job)
}

func int64Ptr(v int64) *int64 { return &v }
func strPtr(s string) *string { return &s }

