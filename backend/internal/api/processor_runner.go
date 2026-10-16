package api

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ProcessorResult struct {
	ProcessingID     string `json:"processing_id"`
	ProcessorVersion string `json:"processor_version"`
	ContractVersion  string `json:"contract_version"`
	Status           string `json:"status"`
}

func RunProcessor(pdfPath string, accountID string) (*ProcessorResult, error) {
	pythonPath := os.Getenv("PROCESSOR_PYTHON")
	if pythonPath == "" {
		if os.PathSeparator == '\\' {
			pythonPath = filepath.Join(os.Getenv("APPDATA"), "FinanceApp", "venv", "Scripts", "python.exe")
		} else {
			pythonPath = filepath.Join(os.Getenv("APPDATA"), "FinanceApp", "venv", "bin", "python")
		}
	}
	scriptPath := os.Getenv("PROCESSOR_SCRIPT")
	if scriptPath == "" {
		scriptPath = filepath.Join(os.Getenv("APPDATA"), "FinanceApp", "processor", "src", "main.py")
	}
	cmd := exec.Command(pythonPath, scriptPath, pdfPath)
	cmd.Env = os.Environ()
	// Ensure artifacts path is available to Python via env
	if artifacts := os.Getenv("ARTIFACTS_PATH"); artifacts != "" {
		cmd.Env = append(cmd.Env, "ARTIFACTS_PATH="+artifacts)
	}
	out, err := cmd.CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf("processor execution failed: %w, output: %s", err, string(out))
	}
	s := strings.TrimSpace(string(out))
	// Strip leading warning lines before JSON
	if idx := strings.IndexByte(s, '{'); idx > 0 {
		s = s[idx:]
	}
	var res ProcessorResult
	if err := json.Unmarshal([]byte(s), &res); err != nil {
		return nil, err
	}
	return &res, nil
}
