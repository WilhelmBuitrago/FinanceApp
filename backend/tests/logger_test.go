package tests

import (
	"financeapp/backend/internal/log"
	"fmt"
	"os"
	"regexp"
	"sync"
	"testing"
)

// TestLogger_ConcurrentWrites pins that concurrent writers (one per HTTP
// handler goroutine) cannot corrupt the log file: every line must arrive
// whole, correctly framed. Run with -race to also detect unsynchronized
// access to the shared file.
func TestLogger_ConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	f, err := log.InitRun(dir)
	if err != nil {
		t.Fatalf("InitRun: %v", err)
	}
	defer func() { _ = f.Close() }()

	const workers, msgs = 20, 25
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < msgs; i++ {
				log.Info(fmt.Sprintf("worker-%d msg-%d", w, i))
			}
		}(w)
	}
	wg.Wait()

	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	lines := regexp.MustCompile(`\r?\n`).Split(string(data), -1)

	// First line is a blank from the RUN header, second is the header itself.
	lineRe := regexp.MustCompile(`^\[\d\d:\d\d:\d\d\] \[INFO\] worker-\d+ msg-\d+$`)
	got := 0
	for _, l := range lines {
		if lineRe.MatchString(l) {
			got++
		}
	}
	want := workers * msgs
	if got != want {
		t.Fatalf("corrupted log: got %d intact lines out of %d", got, want)
	}
}
