package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "modernc.org/sqlite"

	"financeapp/backend/internal/app"
	"financeapp/backend/internal/config"
	"financeapp/backend/internal/db"
	"financeapp/backend/internal/log"
	"financeapp/backend/internal/repositories"
)

func main() {
	cfg := config.Load()
	logDir := cfg.LogsPath
	f, err := log.InitRun(logDir)
	if err != nil {
		fmt.Printf("log init failed: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	log.Log(f, "Backend starting")
	log.Log(f, fmt.Sprintf("DB path: %s", cfg.DatabasePath))
	if err != nil {
		log.Log(f, fmt.Sprintf("failed to open database: %v", err))
		fmt.Printf("failed to open database: %v\n", err)
		os.Exit(1)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Log(f, fmt.Sprintf("failed to ping database: %v", err))
		fmt.Printf("failed to ping database: %v\n", err)
		os.Exit(1)
	}

	if err := db.Apply(sqlDB); err != nil {
		log.Log(f, fmt.Sprintf("migrations failed: %v", err))
		fmt.Printf("migrations failed: %v\n", err)
		sqlDB.Close()
		os.Exit(1)
	}

	querier := db.New(sqlDB)

	repoDB := &repositories.DB{
		Querier: querier,
	}

	// Build repositories
	// Note: Persistence also builds them, but we construct explicitly to avoid double init
	persistence := repositories.NewPersistence(sqlDB)
	_ = repoDB

	application := app.New(sqlDB, persistence)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start application
	if err := application.Start(ctx); err != nil {
		log.Log(f, fmt.Sprintf("application start error: %v", err))
		fmt.Printf("application start error: %v\n", err)
		os.Exit(1)
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("shutdown signal received")
	log.Log(f, "shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	if err := application.Stop(shutdownCtx); err != nil {
		fmt.Printf("application stop error: %v\n", err)
		os.Exit(1)
	}
}
