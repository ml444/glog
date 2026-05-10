package tests

import (
	"bytes"
	"encoding/json"
	"sync"
	"testing"
	"github.com/ml444/glog"
)

type closeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *closeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *closeBuffer) Close() error { return nil }

func (b *closeBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]byte(nil), b.buf.Bytes()...)
}

func TestLoggerNameIsFormatterScoped(t *testing.T) {
	first := &closeBuffer{}
	second := &closeBuffer{}

	newJSONStreamLogger := func(name string, stream *closeBuffer) *log.Logger {
		t.Helper()
		logger, err := log.NewLogger(&log.Config{
			LoggerName:  name,
			LoggerLevel: log.DebugLevel,
			WorkerConfigList: []*log.WorkerConfig{
				log.NewWorkerConfig(log.PrintLevel, 8).
					SetStreamHandlerConfig(&log.StreamHandlerConfig{Streamer: stream}).
					SetJSONFormatterConfig(&log.JSONFormatterConfig{}),
			},
		})
		if err != nil {
			t.Fatalf("NewLogger(%q): %v", name, err)
		}
		return logger
	}

	firstLogger := newJSONStreamLogger("first-service", first)
	secondLogger := newJSONStreamLogger("second-service", second)

	firstLogger.Info("from first")
	secondLogger.Info("from second")

	if err := firstLogger.Stop(); err != nil {
		t.Fatalf("first Stop: %v", err)
	}
	if err := secondLogger.Stop(); err != nil {
		t.Fatalf("second Stop: %v", err)
	}

	assertService := func(name string, raw []byte) {
		t.Helper()
		var record struct {
			Service string `json:"module"`
		}
		if err := json.Unmarshal(bytes.TrimSpace(raw), &record); err != nil {
			t.Fatalf("decode %q output %q: %v", name, string(raw), err)
		}
		if record.Service != name {
			t.Fatalf("service = %q, want %q", record.Service, name)
		}
	}

	assertService("first-service", first.Bytes())
	assertService("second-service", second.Bytes())
}
