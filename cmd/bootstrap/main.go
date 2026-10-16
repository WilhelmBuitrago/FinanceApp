package main

import (
	"fmt"
	"log"
	"financeapp/backend/internal/bootstrap"
	"os"
)

func main() {
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("FinanceApp bootstrap")
	if err := bootstrap.Check(); err != nil {
		fmt.Printf("Missing dependencies: %v\n", err)
		fmt.Println("Run installer to provision Node, Go and Python.")
		os.Exit(1)
	}
	if err := bootstrap.EnsurePythonEnv(baseDir); err != nil {
		log.Fatalf("Bootstrap failed: %v", err)
	}
	fmt.Println("Bootstrap complete. Backend is ready.")
}
