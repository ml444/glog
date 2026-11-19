//go:build linux || darwin

package handler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

func TestReplacementFileHandlerWaitsForRotationAndRefreshesOldFD(t *testing.T) {
	cfg := newRegressionFileConfig(t)
	cfg.MaxFileSize = 4
	cfg.BackupCount = 16
	fm := formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"})
	first, err := NewFileHandler(cfg, fm, nil)
	if err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	var released sync.Once
	unblock := func() { released.Do(func() { close(release) }) }
	t.Cleanup(func() { unblock(); _ = first.Close() })
	write := func(h *FileHandler, text string) {
		t.Helper()
		if err := h.EmitDurable(context.Background(), &message.Entry{Message: text}); err != nil {
			t.Fatal(err)
		}
	}
	write(first, "a")
	write(first, "b")
	r := first.rotator.(*SizeRotator)
	var pause sync.Once
	r.fs.renameFile = func(from, to string) error {
		err := os.Rename(from, to)
		if err == nil && from == r.filePath {
			pause.Do(func() { close(started); <-release })
		}
		return err
	}
	written := make(chan error, 1)
	go func() { written <- first.EmitDurable(context.Background(), &message.Entry{Message: "c"}) }()
	<-started
	type result struct {
		handler *FileHandler
		err     error
	}
	constructed := make(chan result, 1)
	go func() {
		h, err := NewFileHandler(cfg, fm, nil)
		constructed <- result{handler: h, err: err}
	}()
	select {
	case got := <-constructed:
		unblock()
		if got.handler != nil {
			_ = got.handler.Close()
		}
		t.Fatalf("replacement entered an active rotation: %v", got.err)
	case <-time.After(20 * time.Millisecond):
	}
	unblock()
	if err := <-written; err != nil {
		t.Fatal(err)
	}
	got := <-constructed
	if got.err != nil {
		t.Fatal(got.err)
	}
	second := got.handler
	t.Cleanup(func() { _ = second.Close() })
	write(second, "d")
	write(second, "e") // rotates the FD still cached by the first handler
	write(first, "f")
	active, err := os.ReadFile(first.Path())
	if err != nil || string(active) != "e\nf\n" {
		t.Fatalf("old handler appended through a renamed FD: active=%q error=%v", active, err)
	}
	paths, _ := filepath.Glob(first.Path() + "*")
	var content strings.Builder
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		content.Write(data)
	}
	for _, id := range []string{"a", "b", "c", "d", "e", "f"} {
		if strings.Count(content.String(), id+"\n") != 1 {
			t.Fatalf("record %s missing/duplicated across replacement: %q", id, content.String())
		}
	}
}
