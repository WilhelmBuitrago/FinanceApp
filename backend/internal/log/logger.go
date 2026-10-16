package log

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var runID int
var file *os.File

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
	file = f
	runID++
	ts := time.Now().Format("15:04:05")
	header := fmt.Sprintf("\n=== RUN %d %s %s ===\n", runID, date, ts)
	f.WriteString(header)
	return f, nil
}

func Log(f *os.File, msg string) {
	if f == nil {
		return
	}
	ts := time.Now().Format("15:04:05")
	fmt.Fprintf(f, "[%s] %s\n", ts, msg)
}

func LogMsg(msg string) {
	if file == nil {
		return
	}
	ts := time.Now().Format("15:04:05")
	fmt.Fprintf(file, "[%s] %s\n", ts, msg)
}
