package config

import (
	"os"
)

type Config struct {
	DatabasePath string
	Environment  string
}

func Load() Config {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/finance.db"
	}
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	return Config{
		DatabasePath: dbPath,
		Environment:  env,
	}
}
