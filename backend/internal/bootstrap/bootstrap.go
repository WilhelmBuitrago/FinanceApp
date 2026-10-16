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
	PythonURL     = "https://www.python.org/ftp/python/3.12.10/python-3.12.10-embed-amd64.zip"
	PythonSHA256  = "REPLACE_WITH_OFFICIAL_SHA256"
	GetPipURL     = "https://bootstrap.pypa.io/get-pip.py"
)

func AppDataBase() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "FinanceApp")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "AppData", "Roaming", "FinanceApp")
}

func EnsurePythonRuntime() error {
	base := AppDataBase()
	venvPython := filepath.Join(base, "venv", "Scripts", "python.exe")
	if runtime.GOOS != "windows" {
		venvPython = filepath.Join(base, "venv", "bin", "python")
	}
	if ok, _ := pythonVersionMatches(venvPython, PythonVersion); ok {
		if markerOK(base) {
			return nil
		}
	}
	if err := os.MkdirAll(base, 0755); err != nil {
		return err
	}
	runtimeDir := filepath.Join(base, "python-runtime", PythonVersion)
	if err := ensureRuntime(runtimeDir); err != nil {
		return fmt.Errorf("runtime ensure: %w", err)
	}
	if err := createVenv(base, runtimeDir); err != nil {
		return fmt.Errorf("venv create: %w", err)
	}
	if err := installRequirements(venvPython); err != nil {
		return fmt.Errorf("requirements install: %w", err)
	}
	if err := verifyImports(venvPython); err != nil {
		return err
	}
	writeMarker(base)
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
	return strings.Contains(string(data), PythonVersion)
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
	if err := downloadFile(PythonURL, zipPath); err != nil {
		return err
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

func verifyImports(venvPython string) error {
	cmd := exec.Command(venvPython, "-c", "import pdfplumber, pydantic, fitz")
	return cmd.Run()
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
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
