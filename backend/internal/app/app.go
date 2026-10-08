package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"financeapp/backend/internal/api"
	"financeapp/backend/internal/repositories"
)

type Application struct {
	DB           *sql.DB
	Persistence  *repositories.Persistence
	ImportJobRepo *repositories.ImportJobRepo
	DocRepo      *repositories.DocumentRepository
	Server       *http.Server
}

func New(db *sql.DB, persistence *repositories.Persistence) *Application {
	return &Application{
		DB:           db,
		Persistence:  persistence,
		ImportJobRepo: repositories.NewImportJobRepo(db),
		DocRepo:      persistence.DocumentRepo,
	}
}

func (a *Application) Start(ctx context.Context) error {
	if err := a.HealthCheck(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	router := api.NewRouter(a.DocRepo, a.ImportJobRepo, a.Persistence.AccountRepo, a.Persistence.MovementRepo, a.Persistence.FinancialEventRepo, a.Persistence.SourceEntryRepo)
	a.Server = &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		fmt.Println("HTTP server listening on :8080")
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
