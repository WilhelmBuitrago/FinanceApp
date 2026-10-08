package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
	"github.com/google/uuid"
)

type ImportService struct {
	persistence   *repositories.Persistence
	importJobRepo *repositories.ImportJobRepo
}

func NewImportService(p *repositories.Persistence, ijRepo *repositories.ImportJobRepo) *ImportService {
	return &ImportService{persistence: p, importJobRepo: ijRepo}
}

func (s *ImportService) Import(ctx context.Context, jobID string) error {
	job, err := s.importJobRepo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("get job: %w", err)
	}
	if job.Status != "VALIDATING" && job.Status != "PROCESSING" {
		return fmt.Errorf("job status %s not importable", job.Status)
	}
	if job.ProcessingID == nil || *job.ProcessingID == "" {
		return fmt.Errorf("job missing processing_id")
	}
	processingID := *job.ProcessingID

	docRepo := repositories.NewDocumentRepository(&repositories.DB{Querier: s.persistence.Querier})
	doc, err := docRepo.GetByID(ctx, job.DocumentID)
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}
	if doc.AccountID == nil || *doc.AccountID == "" {
		return fmt.Errorf("document account_id is required for importing")
	}
	accountID := *doc.AccountID

	parsedPath := fmt.Sprintf("C:\\Users\\wilhe\\Documents\\FinanceApp\\data\\artifacts\\%s\\parsed.json", processingID)
	data, err := os.ReadFile(parsedPath)
	if err != nil {
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
		Validation struct {
			Valid bool `json:"valid"`
		} `json:"validation"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("unmarshal parsed: %w", err)
	}
	// Validation check
	manifestPath := fmt.Sprintf("C:\\Users\\wilhe\\Documents\\FinanceApp\\data\\artifacts\\%s\\manifest.json", processingID)
	mData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	var manifest struct {
		Validation struct {
			Valid bool `json:"valid"`
		} `json:"validation"`
	}
	if err := json.Unmarshal(mData, &manifest); err != nil {
		return fmt.Errorf("unmarshal manifest: %w", err)
	}
	if !manifest.Validation.Valid {
		return fmt.Errorf("validation not valid")
	}

	// Begin transaction
	tx, err := s.persistence.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	txQ := db.New(tx)
	repoDB := &repositories.DB{Querier: txQ}
	seRepo := repositories.NewSourceEntryRepository(repoDB)
	feRepo := repositories.NewFinancialEventRepository(repoDB)
	mvRepo := repositories.NewMovementRepository(repoDB)

	now := time.Now().UTC().Format(time.RFC3339)

	// Prevent re-import
	job.Status = "IMPORTING"
	if _, err := s.importJobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("update job to importing: %w", err)
	}

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
		if _, err := mvRepo.Create(ctx, mv, []domain.MovementComponent{}); err != nil {
			return fmt.Errorf("create movement: %w", err)
		}
	}

	// Update job to SUCCESS within same tx
	_, err = tx.ExecContext(ctx, `UPDATE import_jobs SET status=?, finished_at=? WHERE id=?`, "SUCCESS", now, jobID)
	if err != nil {
		return fmt.Errorf("update job success: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
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
