package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"financeapp/backend/internal/config"
	"financeapp/backend/internal/log"
	"financeapp/backend/internal/repositories"
)

// router bundles the dependencies every HTTP handler needs. Handlers are
// methods grouped by domain in the handlers_*.go files.
type router struct {
	cfg             config.Config
	docRepo         *repositories.DocumentRepository
	importRepo      *repositories.ImportJobRepo
	accountRepo     *repositories.AccountRepository
	movementRepo    *repositories.MovementRepository
	eventRepo       *repositories.FinancialEventRepository
	sourceEntryRepo *repositories.SourceEntryRepository
	categoryRepo    *repositories.CategoryRepository
	movements       movementWriter
	transfers       transferWriter
	reviews         reviewResolver
}

// ImportRunner owns the transactional staging of a processed document.
// Same rule as movementWriter: api cannot import app, so the transaction
// owner is injected.
type ImportRunner interface {
	Import(ctx context.Context, jobID string) error
}

// ProcessorRunnerFunc runs the Python processor over a saved PDF. It
// carries the request context so a cancelled upload cancels the run.
type ProcessorRunnerFunc func(ctx context.Context, pdfPath, accountID string) (*ProcessorResult, error)

// NewRouter wires every route to its handler. Health and upload stay here;
// the rest live in the handlers_*.go files by domain.
func NewRouter(cfg config.Config, docRepo *repositories.DocumentRepository, importRepo *repositories.ImportJobRepo, accountRepo *repositories.AccountRepository, movementRepo *repositories.MovementRepository, eventRepo *repositories.FinancialEventRepository, sourceEntryRepo *repositories.SourceEntryRepository, categoryRepo *repositories.CategoryRepository, movements movementWriter, transfers transferWriter, reviews reviewResolver, imports ImportRunner, run ProcessorRunnerFunc) http.Handler {
	rt := &router{
		cfg:             cfg,
		docRepo:         docRepo,
		importRepo:      importRepo,
		accountRepo:     accountRepo,
		movementRepo:    movementRepo,
		eventRepo:       eventRepo,
		sourceEntryRepo: sourceEntryRepo,
		categoryRepo:    categoryRepo,
		movements:       movements,
		transfers:       transfers,
		reviews:         reviews,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", rt.health)
	mux.HandleFunc("POST /documents/upload", NewUploadHandler(cfg, docRepo, importRepo, accountRepo, run, imports).Handle)
	mux.HandleFunc("GET /institutions", rt.institutions)
	mux.HandleFunc("POST /accounts", rt.createAccount)
	mux.HandleFunc("GET /accounts", rt.listAccounts)
	mux.HandleFunc("GET /accounts/{id}", rt.getAccount)
	mux.HandleFunc("GET /accounts/{id}/movements", rt.listAccountMovements)
	mux.HandleFunc("GET /accounts/{id}/documents", rt.listAccountDocuments)
	mux.HandleFunc("GET /movements", rt.listMovements)
	mux.HandleFunc("POST /movements", rt.createMovement)
	mux.HandleFunc("GET /movements/recent", rt.recentMovements)
	mux.HandleFunc("PUT /movements/{id}/category", rt.setMovementCategory)
	mux.HandleFunc("DELETE /movements/{id}", rt.deleteMovement)
	mux.HandleFunc("DELETE /accounts/{id}/movements/{mid}", rt.deleteAccountMovement)
	mux.HandleFunc("GET /documents", rt.listDocuments)
	mux.HandleFunc("GET /documents/recent", rt.recentDocuments)
	mux.HandleFunc("GET /documents/{id}", rt.getDocument)
	mux.HandleFunc("GET /financial-events/{id}", rt.getFinancialEvent)
	mux.HandleFunc("GET /source-entries/{id}", rt.getSourceEntry)
	mux.HandleFunc("GET /dashboard/cash-flow", rt.cashFlow)
	mux.HandleFunc("GET /dashboard/income-distribution", rt.incomeDistribution)
	mux.HandleFunc("GET /dashboard/expense-distribution", rt.expenseDistribution)
	mux.HandleFunc("GET /categories", rt.listCategories)
	mux.HandleFunc("POST /categories", rt.createCategory)
	mux.HandleFunc("PUT /categories/{id}", rt.updateCategory)
	mux.HandleFunc("DELETE /categories/{id}", rt.deleteCategory)
	mux.HandleFunc("DELETE /admin/purge", rt.adminPurge)
	mux.HandleFunc("GET /review/entries", rt.listReviewEntries)
	mux.HandleFunc("POST /review/entries/{id}/resolve", rt.resolveReviewEntry)
	mux.HandleFunc("POST /review/entries/{id}/discard", rt.discardReviewEntry)
	mux.HandleFunc("POST /review/entries/{id}/resolve-period", rt.resolveReviewPeriod)
	mux.HandleFunc("GET /jobs/", rt.getJobByPath)
	mux.HandleFunc("GET /jobs/{id}/parsed", rt.jobParsed)
	mux.HandleFunc("GET /jobs/{id}/manifest", rt.jobManifest)

	// Logging middleware
	loggedMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Info(fmt.Sprintf("HTTP %s %s start", r.Method, r.URL.Path))
		rw := &responseWriter{w, http.StatusOK}
		mux.ServeHTTP(rw, r)
		duration := time.Since(start)
		log.Info(fmt.Sprintf("HTTP %s %s %d %s", r.Method, r.URL.Path, rw.status, duration))
	})
	return corsMiddleware(loggedMux)
}

func (rt *router) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encodeJSON(w, map[string]string{"status": "ok"})
}
