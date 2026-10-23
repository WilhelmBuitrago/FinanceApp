package tests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"financeapp/backend/internal/api"
	"financeapp/backend/internal/app"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/domain"
	"financeapp/backend/internal/repositories"
)

// newUploadTestRouter builds the real router over a temporary migrated
// database with isolated paths, an injected import owner and an injected
// processor runner, so the upload flow is exercised without a real PDF.
func newUploadTestRouter(t *testing.T, imports api.ImportRunner, run api.ProcessorRunnerFunc) (http.Handler, *sql.DB) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("DATABASE_PATH", filepath.Join(tmp, "upload-test.db"))
	t.Setenv("DOCUMENTS_PATH", filepath.Join(tmp, "documents"))
	t.Setenv("ARTIFACTS_PATH", filepath.Join(tmp, "artifacts"))
	t.Setenv("LOGS_PATH", filepath.Join(tmp, "logs"))
	sqlDB, err := sql.Open("sqlite", filepath.Join(tmp, "upload-test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Apply(sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	repoDB := &repositories.DB{Querier: db.New(sqlDB)}
	cfg := config.Load()
	if imports == nil {
		imports = app.NewImportService(cfg,
			&repositories.Persistence{DB: sqlDB, Querier: db.New(sqlDB)},
			repositories.NewImportJobRepo(&repositories.DB{Querier: db.New(sqlDB)}))
	}
	router := api.NewRouter(cfg,
		repositories.NewDocumentRepository(repoDB),
		repositories.NewImportJobRepo(&repositories.DB{Querier: db.New(sqlDB)}),
		repositories.NewAccountRepository(repoDB),
		repositories.NewMovementRepository(repoDB),
		repositories.NewFinancialEventRepository(repoDB),
		repositories.NewSourceEntryRepository(repoDB),
		repositories.NewCategoryRepository(repoDB),
		app.NewMovementService(sqlDB),
		app.NewReviewService(sqlDB),
		imports,
		run)
	return router, sqlDB
}

func seedNuAccount(t *testing.T, router http.Handler) {
	t.Helper()
	acc := `{"id":"acc-up","user_id":"u","name":"Nu","account_type":"BANK_ACCOUNT","currency":"COP","status":"ACTIVE","institution":"NU"}`
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader([]byte(acc)))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("setup account: expected %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func newUploadRequest(t *testing.T, pdfBytes []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("account_id", "acc-up")
	_ = mw.WriteField("entity", "NU")
	fw, err := mw.CreateFormFile("file", "extracto.pdf")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(pdfBytes); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/documents/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

// parsedFixture is what the processor leaves on disk: two movements and one
// parked for human review.
const parsedFixture = `{
  "metadata": {"titular": "Test User", "cc": "123", "numero_cuenta": "456", "periodo": "01 - 31 MAR 2026"},
  "summary": {},
  "movements": [
    {"date_original": "01 mar", "date_normalized": "2026-03-01", "description_original": "uno", "amount_original": "+$700.000,00", "amount_normalized": 700000, "page_number": 1, "sequence": 1},
    {"date_original": "02 mar", "date_normalized": "2026-03-02", "description_original": "dos", "amount_original": "-$50.000,00", "amount_normalized": -50000, "page_number": 1, "sequence": 2}
  ],
  "review_required": [
    {"date_original": "03 mar", "description_original": "raro", "amount_original": "+$???", "page_number": 2, "reason": "unparseable amount: no digits"}
  ]
}`

// stubProcessorRunner writes the parsed fixture to the artifacts directory,
// mimicking what the real processor leaves on disk.
func stubProcessorRunner(processingID string) api.ProcessorRunnerFunc {
	return func(pdfPath string, accountID string) (*api.ProcessorResult, error) {
		base := os.Getenv("ARTIFACTS_PATH")
		dir := filepath.Join(base, processingID)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(dir, "parsed.json"), []byte(parsedFixture), 0644); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"validation":{"valid":true}}`), 0644); err != nil {
			return nil, err
		}
		return &api.ProcessorResult{ProcessingID: processingID, Status: "completed"}, nil
	}
}

// TestUpload_DuplicatePdf_ReturnsConflict pins S3: the same file twice is a
// 409 with the existing identifiers, never a 500 and never a duplicate row.
func TestUpload_DuplicatePdf_ReturnsConflict(t *testing.T) {
	router, sqlDB := newUploadTestRouter(t, nil, nil)
	defer func() { _ = sqlDB.Close() }()
	seedNuAccount(t, router)

	pdfBytes := []byte("%PDF-1.4 duplicate probe contents")
	sum := sha256.Sum256(pdfBytes)
	docID := fmt.Sprintf("doc_%x", sum[:4])

	// Seed the document as if it had been imported before.
	docRepo := repositories.NewDocumentRepository(&repositories.DB{Querier: db.New(sqlDB)})
	inst := "NU"
	hashHex := fmt.Sprintf("%x", sum)
	if _, err := docRepo.Create(context.Background(), domain.Document{
		ID:           docID,
		AccountID:    strPtr("acc-up"),
		DocumentType: "PDF",
		Institution:  &inst,
		FileHash:     &hashHex,
		SourceUri:    strPtr("/tmp/old.pdf"),
		Status:       "IMPORTED",
		CreatedAt:    "2026-10-22T00:00:00Z",
		UpdatedAt:    "2026-10-22T00:00:00Z",
	}); err != nil {
		t.Fatalf("seed document: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, newUploadRequest(t, pdfBytes))
	if rr.Code != http.StatusConflict {
		t.Fatalf("re-upload: expected %d, got %d: %s", http.StatusConflict, rr.Code, rr.Body.String())
	}
	var body struct {
		Error      string `json:"error"`
		DocumentID string `json:"document_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode 409 body: %v", err)
	}
	if body.DocumentID != docID {
		t.Fatalf("409 must name the existing document: want %s got %s", docID, body.DocumentID)
	}
}

// TestUpload_StagesThroughService pins S1, S2 and S5: the handler delegates
// to the import owner; movements land in one transaction, parser-parked
// items reach the review queue, and the job closes SUCCESS.
func TestUpload_StagesThroughService(t *testing.T) {
	router, sqlDB := newUploadTestRouter(t, nil, stubProcessorRunner("proc-up-1"))
	defer func() { _ = sqlDB.Close() }()
	seedNuAccount(t, router)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, newUploadRequest(t, []byte("%PDF-1.4 fresh contents")))
	if rr.Code != http.StatusCreated {
		t.Fatalf("upload: expected %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}
	var resp struct {
		DocumentID string `json:"document_id"`
		JobID      string `json:"job_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode upload body: %v", err)
	}
	if resp.DocumentID == "" || resp.JobID == "" {
		t.Fatalf("upload must return document and job ids: %s", rr.Body.String())
	}

	var movements int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM movements`).Scan(&movements); err != nil {
		t.Fatalf("count movements: %v", err)
	}
	if movements != 2 {
		t.Fatalf("expected 2 staged movements, got %d", movements)
	}

	var reviewItems int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM review_items WHERE kind = 'EXTRACTION'`).Scan(&reviewItems); err != nil {
		t.Fatalf("count review items: %v", err)
	}
	if reviewItems != 1 {
		t.Fatalf("parser-parked movement must reach the review queue: expected 1 EXTRACTION item, got %d", reviewItems)
	}

	var jobStatus string
	if err := sqlDB.QueryRow(`SELECT status FROM import_jobs WHERE id = ?`, resp.JobID).Scan(&jobStatus); err != nil {
		t.Fatalf("read job: %v", err)
	}
	if jobStatus != "SUCCESS" {
		t.Fatalf("job must close SUCCESS after a clean import, got %s", jobStatus)
	}
}

// TestUpload_ServiceFailure_ClosesJobFailed pins S5: a failing import owner
// marks the job FAILED and reports 500, never a phantom 201.
type failingImportRunner struct{}

func (failingImportRunner) Import(context.Context, string) error {
	return fmt.Errorf("staging exploded")
}

func TestUpload_ServiceFailure_ClosesJobFailed(t *testing.T) {
	router, sqlDB := newUploadTestRouter(t, failingImportRunner{}, stubProcessorRunner("proc-up-2"))
	defer func() { _ = sqlDB.Close() }()
	seedNuAccount(t, router)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, newUploadRequest(t, []byte("%PDF-1.4 failure probe")))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d when the import owner fails, got %d: %s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
	var jobStatus string
	if err := sqlDB.QueryRow(`SELECT status FROM import_jobs ORDER BY created_at DESC LIMIT 1`).Scan(&jobStatus); err != nil {
		t.Fatalf("read job: %v", err)
	}
	if jobStatus != "FAILED" {
		t.Fatalf("job must close FAILED when staging fails, got %s", jobStatus)
	}
}
