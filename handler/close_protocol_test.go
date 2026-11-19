package handler

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

func TestFileHandlerRejectsEntriesAfterClose(t *testing.T) {
	for _, strategy := range []BackpressureStrategy{
		BackpressureStrategyBlock,
		BackpressureStrategyDrop,
		BackpressureStrategyTimeout,
		BackpressureStrategySample,
	} {
		cfg := newRegressionFileConfig(t)
		cfg.Backpressure = BackpressureConfig{Strategy: strategy, Timeout: time.Hour, SampleRate: 1}
		fm := formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"})
		fileHandler, err := NewFileHandler(cfg, fm, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := fileHandler.Close(); err != nil {
			t.Fatal(err)
		}
		if err := fileHandler.Emit(&message.Entry{Message: "after close"}); !errors.Is(err, ErrHandlerClosed) {
			t.Fatalf("strategy %d: Emit error = %v, want ErrHandlerClosed", strategy, err)
		}
	}
}

func TestFileHandlerCloseReleasesBlockedEmitAndDrainsAcceptedEntries(t *testing.T) {
	for _, strategy := range []BackpressureStrategy{BackpressureStrategyBlock, BackpressureStrategyTimeout} {
		cfg := newRegressionFileConfig(t)
		cfg.BufferSize = 1
		cfg.Backpressure = BackpressureConfig{Strategy: strategy, Timeout: time.Hour}
		fm := &callerBlockingFormatter{started: make(chan struct{}), release: make(chan struct{})}
		fileHandler, err := NewFileHandler(cfg, fm, nil)
		if err != nil {
			t.Fatal(err)
		}
		var releaseOnce sync.Once
		release := func() { releaseOnce.Do(func() { close(fm.release) }) }
		t.Cleanup(func() {
			release()
			_ = fileHandler.Close()
		})
		if err := fileHandler.Emit(&message.Entry{Message: "first"}); err != nil {
			t.Fatal(err)
		}
		select {
		case <-fm.started:
		case <-time.After(time.Second):
			t.Fatal("formatter did not start")
		}
		if err := fileHandler.Emit(&message.Entry{Message: "second"}); err != nil {
			t.Fatal(err)
		}
		emitDone := make(chan error, 1)
		go func() { emitDone <- fileHandler.Emit(&message.Entry{Message: "blocked"}) }()
		select {
		case err := <-emitDone:
			t.Fatalf("Emit did not block: %v", err)
		case <-time.After(10 * time.Millisecond):
		}
		closeDone := make(chan error, 1)
		go func() { closeDone <- fileHandler.Close() }()
		select {
		case err := <-emitDone:
			if !errors.Is(err, ErrHandlerClosed) {
				t.Fatalf("blocked Emit error = %v, want ErrHandlerClosed", err)
			}
		case <-time.After(time.Second):
			t.Fatal("Close did not release blocked Emit")
		}
		select {
		case <-closeDone:
			t.Fatal("Close returned before accepted entries were formatted")
		default:
		}
		release()
		select {
		case err := <-closeDone:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("Close did not finish")
		}
		content, err := os.ReadFile(filepath.Join(cfg.FileDir, "app.log"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(string(content)) != "first\nsecond" {
			t.Fatalf("file content = %q, want both accepted entries", content)
		}
	}
}
