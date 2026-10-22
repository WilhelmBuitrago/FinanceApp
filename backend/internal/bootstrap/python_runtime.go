package bootstrap

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"financeapp/backend/internal/log"
)

// PythonVersion is the runtime version prepared by the installer (via
// resources/prepare-python.ps1) and recorded in the runtime marker.
const PythonVersion = "3.13.17"

// EnsurePythonRuntime validates the Python runtime marker left by the
// installer. It never downloads or installs anything: preparing the runtime
// is the installer's job (resources/prepare-python.ps1).
func EnsurePythonRuntime() error {
	log.Info("EnsurePythonRuntime start")
	base := ProgramDataBase()
	markerPath := filepath.Join(base, ".python-ready")
	log.Info(fmt.Sprintf("Reading marker %s", markerPath))
	data, err := os.ReadFile(markerPath)
	if err != nil {
		log.Error(fmt.Sprintf("marker missing: %v", err))
		return fmt.Errorf("python runtime not prepared: marker missing in %s. Run installer", base)
	}
	content := string(data)
	fields := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			fields[key] = val
		}
	}
	// Required fields
	for _, k := range []string{"version", "requirements_hash", "identifier", "venv_path", "marker_format"} {
		if _, ok := fields[k]; !ok {
			log.Error(fmt.Sprintf("marker incomplete missing %s", k))
			return fmt.Errorf("python runtime marker incomplete: missing %s", k)
		}
	}
	log.Info(fmt.Sprintf("Marker fields ok: version=%s identifier=%s", fields["version"], fields["identifier"]))
	if fields["version"] != PythonVersion {
		log.Error("python runtime version mismatch in marker")
		return fmt.Errorf("python runtime version mismatch in marker")
	}
	if fields["marker_format"] != "1" {
		log.Error("unsupported marker format")
		return fmt.Errorf("unsupported marker format")
	}
	// Validate identifier matches version+hash
	expectedId := PythonVersion + "-" + fields["requirements_hash"]
	if fields["identifier"] != expectedId {
		log.Error(fmt.Sprintf("identifier mismatch: expected %s got %s", expectedId, fields["identifier"]))
		return fmt.Errorf("identifier mismatch: expected %s got %s", expectedId, fields["identifier"])
	}
	log.Info("Identifier validated")
	// Validate requirements_hash against distributed requirements-lock.txt
	reqPath := os.Getenv("PROCESSOR_REQUIREMENTS")
	if reqPath == "" {
		log.Error("PROCESSOR_REQUIREMENTS not set")
		return fmt.Errorf("PROCESSOR_REQUIREMENTS not set, cannot validate requirements hash")
	}
	hash, err := fileSHA256(reqPath)
	if err != nil {
		log.Error(fmt.Sprintf("cannot read requirements-lock: %v", err))
		return fmt.Errorf("cannot read requirements-lock: %w", err)
	}
	if !strings.EqualFold(hash, fields["requirements_hash"]) {
		log.Error("requirements hash mismatch")
		return fmt.Errorf("requirements hash mismatch")
	}
	log.Info("Requirements hash validated")
	venvPathRaw := fields["venv_path"]
	if !filepath.IsAbs(venvPathRaw) {
		log.Error("venv_path must be absolute")
		return fmt.Errorf("venv_path must be absolute")
	}
	absBase, err := filepath.Abs(base)
	if err != nil {
		return fmt.Errorf("cannot resolve runtime base path %s: %w", base, err)
	}
	absVenv, err := filepath.Abs(venvPathRaw)
	if err != nil {
		return fmt.Errorf("cannot resolve venv path %s: %w", venvPathRaw, err)
	}
	expectedVenv := filepath.Join(absBase, "venvs", fields["identifier"])
	if absVenv != expectedVenv {
		log.Error(fmt.Sprintf("venv_path does not match expected %s", expectedVenv))
		return fmt.Errorf("venv_path does not match expected versioned directory")
	}
	log.Info(fmt.Sprintf("venv path validated: %s", venvPathRaw))
	venvPython := filepath.Join(venvPathRaw, "Scripts", "python.exe")
	if runtime.GOOS != "windows" {
		venvPython = filepath.Join(venvPathRaw, "bin", "python")
	}
	// Verify python version matches
	log.Info(fmt.Sprintf("Checking python version at %s", venvPython))
	if ok, err := pythonVersionMatches(venvPython, PythonVersion); err != nil || !ok {
		log.Error(fmt.Sprintf("python runtime version mismatch or not found: %s", venvPython))
		return fmt.Errorf("python runtime version mismatch or not found: %s", venvPython)
	}
	log.Info("Python version matches")
	// Verify imports
	log.Info("Verifying python imports")
	if err := verifyImports(venvPython); err != nil {
		log.Error(fmt.Sprintf("python imports verification failed: %v", err))
		return fmt.Errorf("python imports verification failed: %w", err)
	}
	log.Info("Python imports verified")
	// Verify exact versions
	log.Info("Verifying dependency versions")
	if err := verifyVersions(venvPython); err != nil {
		log.Error(fmt.Sprintf("dependency version mismatch: %v", err))
		return fmt.Errorf("dependency version mismatch: %w", err)
	}
	log.Info("EnsurePythonRuntime success")
	return nil
}

func pythonVersionMatches(path, want string) (bool, error) {
	cmd := exec.Command(path, "--version")
	withNoWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	ver := strings.TrimSpace(string(out))
	return strings.Contains(ver, want), nil
}

func verifyImports(venvPython string) error {
	cmd := exec.Command(venvPython, "-c", "import pdfplumber, pydantic, fitz")
	withNoWindow(cmd)
	return cmd.Run()
}

func verifyVersions(venvPython string) error {
	cmd := exec.Command(venvPython, "-c", "import pdfplumber, pydantic, fitz; import sys; print(pdfplumber.__version__); print(pydantic.__version__); import fitz; print(fitz.__version__)")
	withNoWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(strings.ReplaceAll(string(out), "\r\n", "\n")), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	if len(lines) < 3 {
		return fmt.Errorf("unexpected version output")
	}
	if lines[0] != "0.11.10" {
		return fmt.Errorf("pdfplumber version %s != 0.11.10", lines[0])
	}
	if lines[1] != "2.11.7" {
		return fmt.Errorf("pydantic version %s != 2.11.7", lines[1])
	}
	if lines[2] != "1.26.3" {
		return fmt.Errorf("pymupdf version %s != 1.26.3", lines[2])
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func withNoWindow(cmd *exec.Cmd) {
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		}
	}
}
