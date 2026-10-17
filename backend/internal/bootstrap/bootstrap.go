package bootstrap

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Dependencies struct {
	Node    string
	Go      string
	Python  string
}

func Check() error {
	if err := checkCommand("node", "--version"); err != nil {
		return fmt.Errorf("node not found: %w", err)
	}
	if err := checkCommand("go", "version"); err != nil {
		return fmt.Errorf("go not found: %w", err)
	}
	if err := checkCommand("python", "--version"); err != nil {
		return fmt.Errorf("python not found: %w", err)
	}
	return nil
}

func checkCommand(name, arg string) error {
	cmd := exec.Command(name, arg)
	return cmd.Run()
}

func EnsurePythonEnv(baseDir string) error {
	venvPath := filepath.Join(baseDir, ".venv")
	if _, err := os.Stat(venvPath); os.IsNotExist(err) {
		fmt.Println("Creating Python virtual environment...")
		cmd := exec.Command("python", "-m", "venv", venvPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("venv creation failed: %w", err)
		}
	}
	pip := filepath.Join(venvPath, "Scripts", "pip.exe")
	if runtime.GOOS != "windows" {
		pip = filepath.Join(venvPath, "bin", "pip")
	}
	fmt.Println("Installing processor requirements...")
	reqPath := filepath.Join(baseDir, "apps", "processor", "requirements.txt")
	cmd := exec.Command(pip, "install", "-r", reqPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pip install failed: %w", err)
	}
	return nil
}

const (
	PythonVersion = "3.12.10"
	PythonURL     = "https://www.python.org/ftp/python/3.12.10/python-3.12.10-amd64.zip"
	PythonSHA256  = "8649692de846c56a7189d6dae5c322ab20deb1b5908b6f39426b62a36f39415d"
	GetPipURL     = "https://bootstrap.pypa.io/get-pip.py"
)

func AppDataBase() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "FinanceApp")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "AppData", "Roaming", "FinanceApp")
}

func ProgramDataBase() string {
	if progData := os.Getenv("PROGRAMDATA"); progData != "" {
		return filepath.Join(progData, "FinanceApp")
	}
	return filepath.Join(os.Getenv("SystemDrive"), "ProgramData", "FinanceApp")
}

func EnsurePythonRuntime() error {
	base := ProgramDataBase()
	markerPath := filepath.Join(base, ".python-ready")
	data, err := os.ReadFile(markerPath)
	if err != nil {
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
			return fmt.Errorf("python runtime marker incomplete: missing %s", k)
		}
	}
	if fields["version"] != PythonVersion {
		return fmt.Errorf("python runtime version mismatch in marker")
	}
	if fields["marker_format"] != "1" {
		return fmt.Errorf("unsupported marker format")
	}
	// Validate identifier matches version+hash
	expectedId := PythonVersion + "-" + fields["requirements_hash"]
	if fields["identifier"] != expectedId {
		return fmt.Errorf("identifier mismatch: expected %s got %s", expectedId, fields["identifier"])
	}
	// Validate requirements_hash against distributed requirements-lock.txt
	reqPath := os.Getenv("PROCESSOR_REQUIREMENTS")
	if reqPath == "" {
		return fmt.Errorf("PROCESSOR_REQUIREMENTS not set, cannot validate requirements hash")
	}
	hash, err := fileSHA256(reqPath)
	if err != nil {
		return fmt.Errorf("cannot read requirements-lock: %w", err)
	}
	if strings.ToLower(hash) != strings.ToLower(fields["requirements_hash"]) {
		return fmt.Errorf("requirements hash mismatch")
	}
	venvPathRaw := fields["venv_path"]
	if !filepath.IsAbs(venvPathRaw) {
		return fmt.Errorf("venv_path must be absolute")
	}
	absBase, _ := filepath.Abs(base)
	absVenv, _ := filepath.Abs(venvPathRaw)
	expectedVenv := filepath.Join(absBase, "venvs", fields["identifier"])
	if absVenv != expectedVenv {
		return fmt.Errorf("venv_path does not match expected versioned directory")
	}
	venvPython := filepath.Join(venvPathRaw, "Scripts", "python.exe")
	if runtime.GOOS != "windows" {
		venvPython = filepath.Join(venvPathRaw, "bin", "python")
	}
	// Verify python version matches
	if ok, err := pythonVersionMatches(venvPython, PythonVersion); err != nil || !ok {
		return fmt.Errorf("python runtime version mismatch or not found: %s", venvPython)
	}
	// Verify imports
	if err := verifyImports(venvPython); err != nil {
		return fmt.Errorf("python imports verification failed: %w", err)
	}
	// Verify exact versions
	if err := verifyVersions(venvPython); err != nil {
		return fmt.Errorf("dependency version mismatch: %w", err)
	}
	return nil
}

func pythonVersionMatches(path, want string) (bool, error) {
	cmd := exec.Command(path, "--version")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	ver := strings.TrimSpace(string(out))
	return strings.Contains(ver, want), nil
}

func markerOK(base string) bool {
	data, err := os.ReadFile(filepath.Join(base, ".python-ready"))
	if err != nil {
		return false
	}
	content := string(data)
	if !strings.Contains(content, PythonVersion) {
		return false
	}
	// basic format check
	return strings.Contains(content, "marker_format=1")
}

func writeMarker(base string) {
	os.WriteFile(filepath.Join(base, ".python-ready"), []byte(PythonVersion), 0644)
}

func ensureRuntime(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, "python.exe")); err == nil {
		return nil
	}
	tmp := dir + ".tmp"
	if err := os.MkdirAll(tmp, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	zipPath := filepath.Join(tmp, "python.zip")
	// Prefer bundled zip to avoid network on first run
	if bundle := os.Getenv("PYTHON_BUNDLE_PATH"); bundle != "" {
		if _, err := os.Stat(bundle); err == nil {
			// copy bundled zip to temp
			data, err := os.ReadFile(bundle)
			if err != nil {
				return err
			}
			if err := os.WriteFile(zipPath, data, 0644); err != nil {
				return err
			}
		} else {
			// fallback to download
			if err := downloadFile(PythonURL, zipPath); err != nil {
				return err
			}
		}
	} else {
		if err := downloadFile(PythonURL, zipPath); err != nil {
			return err
		}
	}
	if PythonSHA256 != "" && PythonSHA256 != "REPLACE_WITH_OFFICIAL_SHA256" {
		if err := verifySHA256(zipPath, PythonSHA256); err != nil {
			return err
		}
	}
	if err := unzip(zipPath, dir); err != nil {
		return err
	}
	pythonExe := filepath.Join(dir, "python.exe")
	if _, err := os.Stat(filepath.Join(dir, "Scripts", "pip.exe")); err != nil {
		getPipPath := filepath.Join(tmp, "get-pip.py")
		if err := downloadFile(GetPipURL, getPipPath); err != nil {
			return err
		}
		cmd := exec.Command(pythonExe, getPipPath, "--no-warn-script-location")
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("get-pip failed: %w, %s", err, string(out))
		}
	}
	return nil
}

func createVenv(base, runtimeDir string) error {
	pythonExe := filepath.Join(runtimeDir, "python.exe")
	if _, err := os.Stat(pythonExe); err != nil {
		return fmt.Errorf("runtime python missing")
	}
	venvPath := filepath.Join(base, "venv")
	cmd := exec.Command(pythonExe, "-m", "venv", venvPath)
	cmd.Env = os.Environ()
	return cmd.Run()
}

func installRequirements(venvPython string) error {
	cmd := exec.Command(venvPython, "-m", "pip", "install", "--upgrade", "pip")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pip upgrade failed: %w, %s", err, string(out))
	}
	reqPath := os.Getenv("PROCESSOR_REQUIREMENTS")
	wheelsPath := os.Getenv("PYTHON_WHEELS_PATH")
	useOffline := wheelsPath != "" && dirExists(wheelsPath)
	if useOffline {
		args := []string{"-m", "pip", "install", "--no-index", "--find-links", wheelsPath, "--no-warn-script-location"}
		if reqPath != "" {
			args = append(args, "-r", reqPath)
		}
		cmd = exec.Command(venvPython, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("pip install offline failed: %w, %s", err, string(out))
		}
		return nil
	}
	if reqPath == "" {
		// fallback to known pinned packages
		pkgs := []string{
			"pdfplumber==0.11.10",
			"pydantic==2.11.7",
			"pymupdf==1.26.3",
		}
		cmd = exec.Command(venvPython, "-m", "pip", "install", "--no-warn-script-location")
		cmd.Args = append(cmd.Args, pkgs...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("pip install failed: %w, %s", err, string(out))
		}
		return nil
	}
	cmd = exec.Command(venvPython, "-m", "pip", "install", "--no-warn-script-location", "-r", reqPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip install -r failed: %w, %s", err, string(out))
	}
	return nil
}

func verifyImports(venvPython string) error {
	cmd := exec.Command(venvPython, "-c", "import pdfplumber, pydantic, fitz")
	return cmd.Run()
}

func verifyVersions(venvPython string) error {
	cmd := exec.Command(venvPython, "-c", "import pdfplumber, pydantic, fitz; import sys; print(pdfplumber.__version__); print(pydantic.__version__); import fitz; print(fitz.__version__)")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
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

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func verifySHA256(path, want string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if strings.ToLower(sum) != strings.ToLower(want) {
		return fmt.Errorf("sha256 mismatch")
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
