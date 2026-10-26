package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"financeapp/backend/internal/api"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/log"
	"financeapp/backend/internal/repositories"
)

type Application struct {
	Cfg           config.Config
	DB            *sql.DB
	Persistence   *repositories.Persistence
	ImportJobRepo *repositories.ImportJobRepo
	DocRepo       *repositories.DocumentRepository
	ImportService *ImportService
	Movements     *MovementService
	Reviews       *ReviewService
	Server        *http.Server
}

func New(cfg config.Config, db *sql.DB, persistence *repositories.Persistence) *Application {
	return &Application{
		Cfg:           cfg,
		DB:            db,
		Persistence:   persistence,
		ImportJobRepo: persistence.ImportJobRepo,
		Movements:     NewMovementService(db, NewReviewService(db)),
		Reviews:       NewReviewService(db),
		DocRepo:       persistence.DocumentRepo,
		ImportService: NewImportService(cfg, persistence, persistence.ImportJobRepo),
	}
}

func (a *Application) Start(ctx context.Context) error {
	log.Info("Application health check")
	if err := a.HealthCheck(ctx); err != nil {
		log.Error(fmt.Sprintf("health check failed: %v", err))
		return fmt.Errorf("health check failed: %w", err)
	}
	log.Info("Health check ok")
	router := api.NewRouter(a.Cfg, a.DocRepo, a.ImportJobRepo, a.Persistence.AccountRepo, a.Persistence.MovementRepo, a.Persistence.FinancialEventRepo, a.Persistence.SourceEntryRepo, a.Persistence.CategoryRepo, a.Movements, a.Reviews, a.ImportService, api.RunProcessor)
	port := a.Cfg.Port
	addr := ":" + port
	a.Server = &http.Server{
		Addr:    addr,
		Handler: router,
	}
	go func() {
		log.Info(fmt.Sprintf("HTTP server listening on %s", addr))
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("server error: %v", err))
		}
	}()
	log.Info("Application started")
	return nil
}

func (a *Application) Stop(ctx context.Context) error {
	log.Info("Application stopping")
	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			log.Error(fmt.Sprintf("server shutdown error: %v", err))
			return err
		}
		log.Info("HTTP server shutdown")
	}
	if a.DB != nil {
		if err := a.DB.Close(); err != nil {
			log.Error(fmt.Sprintf("db close error: %v", err))
			return err
		}
		log.Info("Database closed")
	}
	log.Info("Application stopped")
	return nil
}

func (a *Application) HealthCheck(ctx context.Context) error {
	if a.DB == nil {
		return fmt.Errorf("db not configured")
	}
	return a.DB.PingContext(ctx)
}
