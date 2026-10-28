package bootstrap

import (
	"os"
	"path/filepath"
)

// ProgramDataBase returns the machine-wide FinanceApp data directory written
// by the installer.
func ProgramDataBase() string {
	if progData := os.Getenv("PROGRAMDATA"); progData != "" {
		return filepath.Join(progData, "FinanceApp")
	}
	return filepath.Join(os.Getenv("SystemDrive"), "ProgramData", "FinanceApp")
}
