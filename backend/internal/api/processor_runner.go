package api

import (
	"bytes"
	"context"
	"encoding/json"
	"financeapp/backend/internal/log"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ProcessorResult struct {
	ProcessingID     string `json:"processing_id"`
	ProcessorVersion string `json:"processor_version"`
	ContractVersion  string `json:"contract_version"`
	Status           string `json:"status"`
}

// ProcessorTimeout bounds one processor run. The interpreter runs as a
// child process: without a bound, a hung one hangs the upload request
// forever and the job sits in PROCESSING until the app is restarted.
// Preparing an environment takes about a minute; this leaves ample
// headroom. Exported so tests can shrink it.
var ProcessorTimeout = 5 * time.Minute

func RunProcessor(ctx context.Context, pdfPath string, accountID string) (*ProcessorResult, error) {
	log.Info(fmt.Sprintf("RunProcessor start pdf=%s account=%s", pdfPath, accountID))
	pythonPath := os.Getenv("PROCESSOR_PYTHON")
	if pythonPath == "" {
		return nil, fmt.Errorf("PROCESSOR_PYTHON not set: the desktop shell passes the prepared runtime path")
	}
	scriptPath := os.Getenv("PROCESSOR_SCRIPT")
	if scriptPath == "" {
		return nil, fmt.Errorf("PROCESSOR_SCRIPT not set: the desktop shell passes the processor entrypoint")
	}
	log.Info(fmt.Sprintf("Processor python=%s script=%s", pythonPath, scriptPath))
	runCtx, cancel := context.WithTimeout(ctx, ProcessorTimeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, pythonPath, scriptPath, pdfPath)
	cmd.Env = os.Environ()
	// Ensure artifacts path is available to Python via env
	if artifacts := os.Getenv("ARTIFACTS_PATH"); artifacts != "" {
		cmd.Env = append(cmd.Env, "ARTIFACTS_PATH="+artifacts)
	}
	// stdout is the JSON contract; stderr is the processor's log channel
	// (it logs by design). Keeping them apart means a log line never
	// corrupts the result, and a failure still carries its diagnostics.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if runCtx.Err() != nil {
			return nil, fmt.Errorf("processor timed out after %s: %w", ProcessorTimeout, runCtx.Err())
		}
		log.Error(fmt.Sprintf("processor execution failed: %v stderr=%s", err, stderr.String()))
		return nil, fmt.Errorf("processor execution failed: %w, stderr: %s", err, stderr.String())
	}
	log.Info("Processor execution ok")
	s := strings.TrimSpace(stdout.String())
	var res ProcessorResult
	if err := json.Unmarshal([]byte(s), &res); err != nil {
		log.Error(fmt.Sprintf("processor JSON unmarshal failed: %v stdout=%s stderr=%s", err, s, stderr.String()))
		return nil, err
	}
	log.Info(fmt.Sprintf("Processor result processing_id=%s status=%s", res.ProcessingID, res.Status))
	return &res, nil
}
