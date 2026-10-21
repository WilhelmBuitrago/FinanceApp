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

func main() {
	cfg := config.Load()
	log.Info("Config loaded")
	// Ensure required directories exist before any operation that may fail
	for _, p := range []string{filepath.Dir(cfg.DatabasePath), cfg.ArtifactsPath, cfg.LogsPath} {
		if err := os.MkdirAll(p, 0755); err != nil {
			fmt.Printf("failed to create directory %s: %v\n", p, err)
			os.Exit(1)
		}
	}
	log.Info(fmt.Sprintf("Directories ensured: %s, %s, %s", filepath.Dir(cfg.DatabasePath), cfg.ArtifactsPath, cfg.LogsPath))
	logDir := cfg.LogsPath
	f, err := log.InitRun(logDir)
	if err != nil {
		fmt.Printf("log init failed: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	log.Log(f, "Backend starting")
	log.Log(f, fmt.Sprintf("DB path: %s", cfg.DatabasePath))

	// Ensure Python runtime before starting services
	log.Info("Checking Python runtime")
	if err := bootstrap.EnsurePythonRuntime(); err != nil {
		log.Error(fmt.Sprintf("Python runtime not ready: %v", err))
		log.Log(f, fmt.Sprintf("Python runtime not ready: %v", err))
		fmt.Printf("Python runtime not ready: %v\n", err)
		os.Exit(1)
	}
	log.Info("Python runtime ready")
	// Ensure required directories exist

	sqlDB, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		log.Error(fmt.Sprintf("failed to open database: %v", err))
		log.Log(f, fmt.Sprintf("failed to open database: %v", err))
		fmt.Printf("failed to open database: %v\n", err)
		os.Exit(1)
	}
	log.Info("Database opened")

	if err := sqlDB.Ping(); err != nil {
		log.Error(fmt.Sprintf("failed to ping database: %v", err))
		log.Log(f, fmt.Sprintf("failed to ping database: %v", err))
		fmt.Printf("failed to ping database: %v\n", err)
		os.Exit(1)
	}
	log.Info("Database ping ok")

	if err := db.Apply(sqlDB); err != nil {
		log.Error(fmt.Sprintf("migrations failed: %v", err))
		log.Log(f, fmt.Sprintf("migrations failed: %v", err))
		fmt.Printf("migrations failed: %v\n", err)
		sqlDB.Close()
		os.Exit(1)
	}
	log.Info("Migrations applied")

	persistence := repositories.NewPersistence(sqlDB)

	application := app.New(cfg, sqlDB, persistence)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("Starting application")
	if err := application.Start(ctx); err != nil {
		log.Error(fmt.Sprintf("application start error: %v", err))
		log.Log(f, fmt.Sprintf("application start error: %v", err))
		fmt.Printf("application start error: %v\n", err)
		os.Exit(1)
	}
	log.Info("Application started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("shutdown signal received")
	log.Info("shutdown signal received")
	log.Log(f, "shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	if err := application.Stop(shutdownCtx); err != nil {
		log.Error(fmt.Sprintf("application stop error: %v", err))
		fmt.Printf("application stop error: %v\n", err)
		os.Exit(1)
	}
}

