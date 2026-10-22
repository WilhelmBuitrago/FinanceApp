package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Guards every write to the shared log file: the HTTP server handles requests
// in concurrent goroutines and all of them log.
var (
	mu    sync.Mutex
	runID int
	file  *os.File
)

func logWrite(msg string) {
	mu.Lock()
	defer mu.Unlock()
	if file == nil {
		return
	}
	ts := time.Now().Format("15:04:05")
	_, _ = fmt.Fprintf(file, "[%s] %s\n", ts, msg)
}

func InitRun(logDir string) (*os.File, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	date := time.Now().Format("02-01-2006")
	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", date))
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	defer mu.Unlock()
	file = f
	runID++
	ts := time.Now().Format("15:04:05")
	header := fmt.Sprintf("\n=== RUN %d %s %s ===\n", runID, date, ts)
	_, _ = f.WriteString(header)
	return f, nil
}

func Info(msg string) {
	logWrite("[INFO] " + msg)
}

func Warn(msg string) {
	logWrite("[WARN] " + msg)
}

func Error(msg string) {
	logWrite("[ERROR] " + msg)
}
