package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/log"
	"financeapp/backend/internal/repositories"

	"github.com/google/uuid"
)

type ImportService struct {
	cfg           config.Config
	persistence   *repositories.Persistence
	importJobRepo *repositories.ImportJobRepo
}

func NewImportService(cfg config.Config, p *repositories.Persistence, ijRepo *repositories.ImportJobRepo) *ImportService {
	return &ImportService{cfg: cfg, persistence: p, importJobRepo: ijRepo}
}

func (s *ImportService) Import(ctx context.Context, jobID string) error {
	log.Info(fmt.Sprintf("ImportService.Import start jobID=%s", jobID))
	job, err := s.importJobRepo.GetByID(ctx, jobID)
	if err != nil {
		log.Error(fmt.Sprintf("get job error: %v", err))
		return fmt.Errorf("get job: %w", err)
	}
	if job.Status != "VALIDATING" && job.Status != "PROCESSING" {
		log.Error(fmt.Sprintf("job status not importable %s", job.Status))
		return fmt.Errorf("job status %s not importable", job.Status)
	}
	if job.ProcessingID == nil || *job.ProcessingID == "" {
		log.Error("job missing processing_id")
		return fmt.Errorf("job missing processing_id")
	}
	processingID := *job.ProcessingID
	log.Info(fmt.Sprintf("processingID=%s", processingID))

	docRepo := repositories.NewDocumentRepository(&repositories.DB{Querier: s.persistence.Querier})
	doc, err := docRepo.GetByID(ctx, job.DocumentID)
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}
	if doc.AccountID == nil || *doc.AccountID == "" {
		return fmt.Errorf("document account_id is required for importing")
	}
	accountID := *doc.AccountID

	log.Info("Import: reading parsed.json")
	parsedPath := filepath.Join(s.cfg.ArtifactsPath, processingID, "parsed.json")
	data, err := os.ReadFile(parsedPath)
	if err != nil {
		log.Error("Import: failed to read parsed.json")
		return fmt.Errorf("read parsed.json: %w", err)
	}
	var parsed struct {
		Metadata  map[string]interface{} `json:"metadata"`
		Summary   map[string]interface{} `json:"summary"`
		Movements []struct {
			DateOriginal        string `json:"date_original"`
			DateNormalized      string `json:"date_normalized"`
			DescriptionOriginal string `json:"description_original"`
			AmountOriginal      string `json:"amount_original"`
			AmountNormalized    int64  `json:"amount_normalized"`
			PageNumber          int    `json:"page_number"`
			Sequence            int    `json:"sequence"`
		} `json:"movements"`
		ReviewRequired []struct {
			DateOriginal        string `json:"date_original"`
			DescriptionOriginal string `json:"description_original"`
			AmountOriginal      string `json:"amount_original"`
			AmountNormalized    int64  `json:"amount_normalized"`
			PageNumber          int    `json:"page_number"`
			Reason              string `json:"reason"`
		} `json:"review_required"`
		DocumentReview []struct {
			RawPeriod string `json:"raw_period"`
			Reason    string `json:"reason"`
		} `json:"document_review"`
		Validation struct {
			Valid bool `json:"valid"`
		} `json:"validation"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		log.Error("Import: unmarshal parsed.json failed")
		return fmt.Errorf("unmarshal parsed: %w", err)
	}
	log.Info(fmt.Sprintf("Import: parsed %d movements", len(parsed.Movements)))
	// Validation check
	log.Info("Import: reading manifest.json")
	manifestPath := filepath.Join(s.cfg.ArtifactsPath, processingID, "manifest.json")
	mData, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Error("Import: failed to read manifest.json")
		return fmt.Errorf("read manifest: %w", err)
	}
	var manifest struct {
		Validation struct {
			Valid bool `json:"valid"`
		} `json:"validation"`
	}
	if err := json.Unmarshal(mData, &manifest); err != nil {
		log.Error("Import: unmarshal manifest failed")
		return fmt.Errorf("unmarshal manifest: %w", err)
	}
	if !manifest.Validation.Valid {
		log.Error("Import: validation not valid")
		return fmt.Errorf("validation not valid")
	}
	log.Info("Import: validation ok")

	// Begin transaction
	log.Info("Import: beginning transaction")
	tx, err := s.persistence.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Error("Import: begin tx failed")
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			log.Warn("Import: transaction rollback")
			_ = tx.Rollback()
		}
	}()

	txQ := db.New(tx)
	repoDB := &repositories.DB{Querier: txQ}
	seRepo := repositories.NewSourceEntryRepository(repoDB)
	feRepo := repositories.NewFinancialEventRepository(repoDB)
	mvRepo := repositories.NewMovementRepository(repoDB)
	reviewItemsRepo := repositories.NewReviewItemsRepository(repoDB)

	now := time.Now().UTC().Format(time.RFC3339)

	// Prevent re-import
	log.Info("Import: updating job status to IMPORTING")
	job.Status = "IMPORTING"
	if _, err := s.importJobRepo.Update(ctx, job); err != nil {
		log.Error("Import: failed to update job status")
		return fmt.Errorf("update job to importing: %w", err)
	}
	log.Info(fmt.Sprintf("Import: processing %d movements", len(parsed.Movements)))

	for _, m := range parsed.Movements {
		seID := uuid.NewString()
		se := domain.SourceEntry{
			ID:                    seID,
			DocumentID:            job.DocumentID,
			RawDate:               strPtr(m.DateOriginal),
			RawDescription:        strPtr(m.DescriptionOriginal),
			RawAmountMinor:        nil,
			NormalizedDate:        strPtr(m.DateNormalized),
			NormalizedDescription: strPtr(m.DescriptionOriginal),
			NormalizedAmountMinor: int64Ptr(m.AmountNormalized),
			PageNumber:            int64Ptr(int64(m.PageNumber)),
			RowNumber:             int64Ptr(int64(m.Sequence)),
			Section:               nil,
			ExtractionID:          strPtr(processingID),
			CreatedAt:             now,
			UpdatedAt:             now,
		}
		if _, err := seRepo.Create(ctx, se); err != nil {
			return fmt.Errorf("create source entry: %w", err)
		}

		feID := uuid.NewString()
		eventType := classifyEventType(m.AmountNormalized)
		fe := domain.FinancialEvent{
			ID:                 feID,
			EventType:          eventType,
			EventDate:          strPtr(m.DateNormalized),
			Status:             "ACTIVE",
			VerificationStatus: "UNVERIFIED",
			Description:        strPtr(m.DescriptionOriginal),
			CategoryID:         nil,
			Confidence:         nil,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if _, err := feRepo.Create(ctx, fe); err != nil {
			return fmt.Errorf("create financial event: %w", err)
		}

		mvID := uuid.NewString()
		mv := domain.Movement{
			ID:               mvID,
			FinancialEventID: feID,
			AccountID:        accountID,
			SourceEntryID:    strPtr(seID),
			MovementDate:     m.DateNormalized,
			AmountMinor:      m.AmountNormalized,
			Description:      strPtr(m.DescriptionOriginal),
			Status:           "ACTIVE",
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if _, err := mvRepo.Create(ctx, mv); err != nil {
			return fmt.Errorf("create movement: %w", err)
		}
	}

	// Movements the parser parked for review: the raw becomes a source entry
	// flagged REVIEW_REQUIRED (a human decides later), never a guessed amount.
	// source_key is stable per document+page+raw so re-importing does not
	// duplicate the queue.
	for i, rv := range parsed.ReviewRequired {
		se := domain.SourceEntry{
			ID:                    fmt.Sprintf("%s_rev_%d", job.DocumentID, i+1),
			DocumentID:            job.DocumentID,
			SourceReference:       strPtr("review:" + rv.Reason),
			RawDescription:        strPtr(rv.DescriptionOriginal),
			RawDate:               strPtr(rv.DateOriginal),
			NormalizedDescription: strPtr(rv.DescriptionOriginal),
			NormalizedAmountMinor: int64Ptr(rv.AmountNormalized),
			PageNumber:            int64Ptr(int64(rv.PageNumber)),
			ExtractionID:          strPtr(processingID),
			ReviewStatus:          "REVIEW_REQUIRED",
			SourceKey:             strPtr(fmt.Sprintf("%s:p%d:%s", job.DocumentID, rv.PageNumber, rv.AmountOriginal)),
			MetadataJson:          strPtr(fmt.Sprintf(`{"reason":%q,"amount_original":%q}`, rv.Reason, rv.AmountOriginal)),
			CreatedAt:             now,
			UpdatedAt:             now,
		}
		if _, err := seRepo.Create(ctx, se); err != nil {
			return fmt.Errorf("create review source entry: %w", err)
		}
		// The same parked movement also lands in the unified human-review
		// queue (kind EXTRACTION) pointing at the source entry: one place for
		// the UI, all three triggers.
		payload, err := json.Marshal(map[string]any{
			"reason":               rv.Reason,
			"amount_original":      rv.AmountOriginal,
			"date_original":        rv.DateOriginal,
			"description_original": rv.DescriptionOriginal,
			"page_number":          rv.PageNumber,
			"document_id":          job.DocumentID,
			"account_id":           accountID,
		})
		if err != nil {
			return fmt.Errorf("marshal review payload: %w", err)
		}
		seID := se.ID
		if err := reviewItemsRepo.Create(ctx, uuid.NewString(), "EXTRACTION", strPtr(string(payload)), &seID, now); err != nil {
			return fmt.Errorf("park review item: %w", err)
		}
	}

	// Document-level review: the statement period could not be read (the
	// silent retry already failed). One item, no movement attached: the
	// human supplies month+year and the parked movement dates are derived
	// against it on resolve.
	for _, dr := range parsed.DocumentReview {
		payload, err := json.Marshal(map[string]any{
			"raw_period":  dr.RawPeriod,
			"reason":      dr.Reason,
			"document_id": job.DocumentID,
			"account_id":  accountID,
		})
		if err != nil {
			return fmt.Errorf("marshal document review payload: %w", err)
		}
		if err := reviewItemsRepo.Create(ctx, uuid.NewString(), reviewKindDocumentPeriod, strPtr(string(payload)), nil, now); err != nil {
			return fmt.Errorf("park document review item: %w", err)
		}
	}

	// Update job to SUCCESS within same tx (repo bound to the transaction so
	// the status change commits or rolls back with the movements).
	jobTxRepo := repositories.NewImportJobRepo(&repositories.DB{Querier: db.New(tx)})
	job.Status = "SUCCESS"
	job.FinishedAt = &now
	if _, err := jobTxRepo.Update(ctx, job); err != nil {
		log.Error("Import: failed to update job to SUCCESS")
		return fmt.Errorf("update job success: %w", err)
	}

	log.Info("Import: committing transaction")
	if commitErr := tx.Commit(); commitErr != nil {
		log.Error(fmt.Sprintf("Import: commit failed: %v", commitErr))
		// Keep the deferred rollback armed: a failed commit may leave an open tx.
		err = commitErr
		return fmt.Errorf("commit: %w", commitErr)
	}
	log.Info("Import complete")
	return nil
}

func classifyEventType(amount int64) string {
	if amount > 0 {
		return "INCOME"
	}
	return "EXPENSE"
}

func strPtr(s string) *string { return &s }
func int64Ptr(i int64) *int64 { return &i }
