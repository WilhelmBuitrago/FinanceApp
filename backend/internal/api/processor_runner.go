package api

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
		pythonPath = `C:\Users\wilhe\miniconda3\envs\agent\python.exe`
	}
	cmd := exec.Command(pythonPath, "../apps/processor/src/main.py", pdfPath)
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
