package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	_ "modernc.org/sqlite"

	"financeapp/backend/internal/app"
	"financeapp/backend/internal/bootstrap"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/log"
	"financeapp/backend/internal/repositories"
)

// fatal records a startup failure in the log and exits. Callers must only use
// it once the log file exists (after InitRun); before that, stdout is the only
// channel left.
func fatal(msg string, err error) {
	log.Error(fmt.Sprintf("%s: %v", msg, err))
	os.Exit(1)
}

func main() {
	cfg := config.Load()
	log.Info("Config loaded")
	// Ensure required directories exist before any operation that may fail
	for _, p := range []string{filepath.Dir(cfg.DatabasePath), cfg.ArtifactsPath, cfg.LogsPath} {
		if err := os.MkdirAll(p, 0755); err != nil {
			// Log not initialized yet: stdout is the only channel available.
			fmt.Printf("failed to create directory %s: %v\n", p, err)
			os.Exit(1)
		}
	}
	log.Info(fmt.Sprintf("Directories ensured: %s, %s, %s", filepath.Dir(cfg.DatabasePath), cfg.ArtifactsPath, cfg.LogsPath))
	f, err := log.InitRun(cfg.LogsPath)
	if err != nil {
		fmt.Printf("log init failed: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = f.Close() }()
	log.Info("Backend starting")
	log.Info(fmt.Sprintf("DB path: %s", cfg.DatabasePath))

	// Ensure Python runtime before starting services
	log.Info("Checking Python runtime")
	if err := bootstrap.EnsurePythonRuntime(); err != nil {
		fatal("Python runtime not ready", err)
	}
	log.Info("Python runtime ready")

	sqlDB, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		fatal("failed to open database", err)
	}
	log.Info("Database opened")

	if err := sqlDB.Ping(); err != nil {
		fatal("failed to ping database", err)
	}
	log.Info("Database ping ok")

	if err := db.Apply(sqlDB); err != nil {
		log.Error(fmt.Sprintf("migrations failed: %v", err))
		_ = sqlDB.Close()
		os.Exit(1)
	}
	log.Info("Migrations applied")

	persistence := repositories.NewPersistence(sqlDB)

	application := app.New(cfg, sqlDB, persistence)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("Starting application")
	if err := application.Start(ctx); err != nil {
		fatal("application start error", err)
	}
	log.Info("Application started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	if err := application.Stop(shutdownCtx); err != nil {
		fatal("application stop error", err)
	}
}
