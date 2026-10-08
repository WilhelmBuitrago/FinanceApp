package api

import (
	"encoding/json"
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
	cmd := exec.Command("python", "../apps/processor/src/main.py", pdfPath, accountID)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
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
