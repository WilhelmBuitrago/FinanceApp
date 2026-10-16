package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"financeapp/backend/internal/api"
	"financeapp/backend/internal/repositories"
)

type Application struct {
	DB           *sql.DB
	Persistence  *repositories.Persistence
	ImportJobRepo *repositories.ImportJobRepo
	DocRepo      *repositories.DocumentRepository
	ImportService *ImportService
	Server       *http.Server
}

func New(db *sql.DB, persistence *repositories.Persistence) *Application {
	importJobRepo := repositories.NewImportJobRepo(db)
	return &Application{
		DB:           db,
		Persistence:  persistence,
		ImportJobRepo: importJobRepo,
		DocRepo:      persistence.DocumentRepo,
		ImportService: NewImportService(persistence, importJobRepo),
	}
}

func (a *Application) Start(ctx context.Context) error {
	if err := a.HealthCheck(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	router := api.NewRouter(a.DocRepo, a.ImportJobRepo, a.Persistence.AccountRepo, a.Persistence.MovementRepo, a.Persistence.FinancialEventRepo, a.Persistence.SourceEntryRepo, a.Persistence.CategoryRepo, func(ctx context.Context, id string) error {
		return a.ImportService.Import(ctx, id)
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port
	a.Server = &http.Server{
		Addr:    addr,
		Handler: router,
	}
	go func() {
		fmt.Printf("HTTP server listening on %s\n", addr)
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("server error: %v\n", err)
		}
	}()
	fmt.Println("Application started")
	return nil
}

func (a *Application) Stop(ctx context.Context) error {
	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			return err
		}
	}
	if a.DB != nil {
		if err := a.DB.Close(); err != nil {
			return err
		}
	}
	fmt.Println("Application stopped")
	return nil
}

func (a *Application) HealthCheck(ctx context.Context) error {
	if a.DB == nil {
		return fmt.Errorf("db not configured")
	}
	return a.DB.PingContext(ctx)
}
