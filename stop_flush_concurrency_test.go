package log

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ml444/glog/message"
)

// collectHandler implements handler.IHandler for tests: records every emitted message.
type collectHandler struct {
	mu    sync.Mutex
	lines []string
}

func (h *collectHandler) Emit(e *message.Entry) error {
	h.mu.Lock()
	h.lines = append(h.lines, e.Message)
	h.mu.Unlock()
	return nil
}

func (h *collectHandler) Close() error { return nil }

func (h *collectHandler) len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.lines)
}

// TestStopFlushesAllConcurrentLogs checks that after heavy concurrent logging,
// Logger.Stop drains the pipeline so the handler receives every message (no silent drops).
func TestStopFlushesAllConcurrentLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping high-volume test in -short")
	}

	const (
		goroutines = 128
		perRoutine = 500
	)
	want := goroutines * perRoutine // 64000

	h := &collectHandler{}
	cfg := &Config{
		LoggerLevel: DebugLevel,
		WorkerConfigList: []*WorkerConfig{
			// Small channel to increase contention between Send and worker drain.
			NewWorkerConfig(PrintLevel, 32).SetHandler(h),
		},
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	var seq int64
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perRoutine; i++ {
				n := atomic.AddInt64(&seq, 1)
				logger.Infof("id:%d", n)
			}
		}()
	}
	wg.Wait()

	if got := int(atomic.LoadInt64(&seq)); got != want {
		t.Fatalf("internal count mismatch: got %d want %d", got, want)
	}

	if err := logger.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	if got := h.len(); got != want {
		t.Fatalf("handler saw %d lines, want %d (Stop should drain all queued logs)", got, want)
	}
}

// TestStopFlushesConcurrentLogsShort is a lighter variant for -short runs.
func TestStopFlushesConcurrentLogsShort(t *testing.T) {
	if !testing.Short() {
		t.Skip("only runs with -short (use TestStopFlushesAllConcurrentLogs for full load)")
	}

	const (
		goroutines = 32
		perRoutine = 50
	)
	want := goroutines * perRoutine

	h := &collectHandler{}
	cfg := &Config{
		LoggerLevel: DebugLevel,
		WorkerConfigList: []*WorkerConfig{
			NewWorkerConfig(PrintLevel, 8).SetHandler(h),
		},
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	var seq int64
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perRoutine; i++ {
				n := atomic.AddInt64(&seq, 1)
				logger.Infof("id:%d", n)
			}
		}()
	}
	wg.Wait()

	if got := int(atomic.LoadInt64(&seq)); got != want {
		t.Fatalf("internal count mismatch: got %d want %d", got, want)
	}
	if err := logger.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if got := h.len(); got != want {
		t.Fatalf("handler saw %d lines, want %d", got, want)
	}
}
