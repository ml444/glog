package log

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/ml444/glog/message"
)

type captureHandler struct {
	entries chan *message.Entry
	closed  uint64
}

func (h *captureHandler) Emit(entry *message.Entry) error {
	h.entries <- entry
	return nil
}

func (h *captureHandler) Close() error {
	atomic.AddUint64(&h.closed, 1)
	return nil
}

func TestNewEnabledAndRoutineIDOption(t *testing.T) {
	h := &captureHandler{entries: make(chan *message.Entry, 1)}
	logger, err := New(
		SetLoggerLevel(InfoLevel),
		SetRecordRoutineID(false),
		SetWorkerConfigs(
			NewWorkerConfig(InfoLevel, 8).
				SetBackpressureStrategy(BackpressureStrategyBlock).
				SetHandler(h),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Stop() }()

	if logger.Enabled(DebugLevel) {
		t.Fatal("DebugLevel unexpectedly enabled")
	}
	if !logger.Enabled(InfoLevel) {
		t.Fatal("InfoLevel unexpectedly disabled")
	}
	logger.Log(InfoLevel, "hello")

	select {
	case entry := <-h.entries:
		if entry.RoutineID != 0 {
			t.Fatalf("RoutineID = %d, want 0", entry.RoutineID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for entry")
	}
}
