package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"financeapp/backend/internal/api"
)

// pythonInterpreter resolves the processor interpreter the repo documents
// (the conda env path). The test skips when no interpreter is available:
// it pins the stdout/stderr contract, not Python itself.
func pythonInterpreter(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("PROCESSOR_PYTHON"); p != "" {
		return p
	}
	candidate := filepath.Join(os.Getenv("LOCALAPPDATA"), "..", "..", "miniconda3", "envs", "agent", "python.exe")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	t.Skip("no PROCESSOR_PYTHON and the documented conda env is absent")
	return ""
}

func writeProcessorScript(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake_processor.py")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return path
}

// TestRunProcessor_StdoutIsJsonEvenWhenStderrLogs pins the contract: the
// result comes from stdout alone. A log line on stderr that contains a
// brace must not corrupt the JSON (with CombinedOutput it did: the naive
// first-brace strip cut into the log garbage).
func TestRunProcessor_StdoutIsJsonEvenWhenStderrLogs(t *testing.T) {
	py := pythonInterpreter(t)
	t.Setenv("PROCESSOR_PYTHON", py)
	t.Setenv("PROCESSOR_SCRIPT", writeProcessorScript(t, `
import sys
sys.stderr.write("2026-10-23 12:00:00 [INFO] process_pdf start {\n")
sys.stderr.write("2026-10-23 12:00:01 [WARNING] period unreadable {\n")
sys.stdout.write('{"processing_id":"proc-contract","processor_version":"0.1.0","contract_version":"1","status":"completed"}\n')
`))

	res, err := api.RunProcessor(context.Background(), `C:\dummy.pdf`, "acc-1")
	if err != nil {
		t.Fatalf("stderr noise must not break the JSON contract: %v", err)
	}
	if res.ProcessingID != "proc-contract" {
		t.Fatalf("processing_id must come from stdout, got %q", res.ProcessingID)
	}
}

// TestRunProcessor_FailureReportsStderr pins that a failing processor still
// surfaces its stderr detail for debugging, and never a fabricated result.
func TestRunProcessor_FailureReportsStderr(t *testing.T) {
	py := pythonInterpreter(t)
	t.Setenv("PROCESSOR_PYTHON", py)
	t.Setenv("PROCESSOR_SCRIPT", writeProcessorScript(t, `
import sys
sys.stderr.write("Traceback (most recent call last):\nValueError: boom-detail-42\n")
sys.exit(1)
`))

	res, err := api.RunProcessor(context.Background(), `C:\dummy.pdf`, "acc-1")
	if err == nil {
		t.Fatal("a failing processor must return an error")
	}
	if res != nil {
		t.Fatal("a failing processor must not return a result")
	}
	if !contains(err.Error(), "boom-detail-42") {
		t.Fatalf("the error must carry the stderr detail for debugging, got: %v", err)
	}
}

// TestRunProcessor_TimeoutKillsHungRun pins that a hung interpreter does
// not hang the upload request: the run is cancelled near the timeout and
// the error says so. The ceiling doubles as the teeth check — without
// the cancellation the run would last the script's full sleep.
func TestRunProcessor_TimeoutKillsHungRun(t *testing.T) {
	py := pythonInterpreter(t)
	t.Setenv("PROCESSOR_PYTHON", py)
	t.Setenv("PROCESSOR_SCRIPT", writeProcessorScript(t, `
import time
time.sleep(30)
`))

	previous := api.ProcessorTimeout
	api.ProcessorTimeout = 300 * time.Millisecond
	defer func() { api.ProcessorTimeout = previous }()

	start := time.Now()
	res, err := api.RunProcessor(context.Background(), `C:\dummy.pdf`, "acc-1")
	if err == nil {
		t.Fatal("a hung processor must not return a result")
	}
	if res != nil {
		t.Fatal("a cancelled run must not return a result")
	}
	if !contains(err.Error(), "timed out") {
		t.Fatalf("the error must report the timeout, got: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("the run must be cancelled near the timeout, took %s", elapsed)
	}
}
