package api

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"financeapp/backend/internal/log"
)

type ProcessorResult struct {
	ProcessingID     string `json:"processing_id"`
	ProcessorVersion string `json:"processor_version"`
	ContractVersion  string `json:"contract_version"`
	Status           string `json:"status"`
}

func RunProcessor(pdfPath string, accountID string) (*ProcessorResult, error) {
	log.Info(fmt.Sprintf("RunProcessor start pdf=%s account=%s", pdfPath, accountID))
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
	log.Info(fmt.Sprintf("Processor python=%s script=%s", pythonPath, scriptPath))
	cmd := exec.Command(pythonPath, scriptPath, pdfPath)
	cmd.Env = os.Environ()
	// Ensure artifacts path is available to Python via env
	if artifacts := os.Getenv("ARTIFACTS_PATH"); artifacts != "" {
		cmd.Env = append(cmd.Env, "ARTIFACTS_PATH="+artifacts)
	}
	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Error(fmt.Sprintf("processor execution failed: %v output=%s", err, string(out)))
		return nil, fmt.Errorf("processor execution failed: %w, output: %s", err, string(out))
	}
	log.Info("Processor execution ok")
	s := strings.TrimSpace(string(out))
	// Strip leading warning lines before JSON
	if idx := strings.IndexByte(s, '{'); idx > 0 {
		s = s[idx:]
	}
	var res ProcessorResult
	if err := json.Unmarshal([]byte(s), &res); err != nil {
		log.Error(fmt.Sprintf("processor JSON unmarshal failed: %v output=%s", err, s))
		return nil, err
	}
	log.Info(fmt.Sprintf("Processor result processing_id=%s status=%s", res.ProcessingID, res.Status))
	return &res, nil
}
