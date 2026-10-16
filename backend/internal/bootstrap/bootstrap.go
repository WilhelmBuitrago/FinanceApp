package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
