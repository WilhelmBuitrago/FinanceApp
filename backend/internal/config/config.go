package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DatabasePath  string
	ArtifactsPath string
	LogsPath      string
	Environment   string
	PythonPath    string
	Port          string
}

func defaultAppData() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "FinanceApp")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "AppData", "Roaming", "FinanceApp")
}

func defaultPath(sub string) string {
	base := defaultAppData()
	if base == "" {
		return ""
	}
	return filepath.Join(base, sub)
}

func Load() Config {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = filepath.Join(defaultPath("data/db"), "finance.db")
	}
	artifacts := os.Getenv("ARTIFACTS_PATH")
	if artifacts == "" {
		artifacts = defaultPath("data/artifacts")
	}
	logs := os.Getenv("LOGS_PATH")
	if logs == "" {
		logs = defaultPath("logs")
	}
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	python := os.Getenv("PROCESSOR_PYTHON")
	if python == "" {
		python = filepath.Join(defaultAppData(), "venv", "bin", "python")
		if os.PathSeparator == '\\' {
			python = filepath.Join(defaultAppData(), "venv", "Scripts", "python.exe")
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	return Config{
		DatabasePath:  dbPath,
		ArtifactsPath: artifacts,
		LogsPath:      logs,
		Environment:   env,
		PythonPath:    python,
		Port:          port,
	}
}
