//go:build linux || darwin

package handler

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

func newDurableTestHandler(t *testing.T) *FileHandler {
	t.Helper()
	cfg := newRegressionFileConfig(t)
	h, err := NewFileHandler(cfg, formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"}), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = h.Close() })
	return h
}

func TestEmitDurableWaitsForSynchronization(t *testing.T) {
	h := newDurableTestHandler(t)
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	h.syncFile = func(file *os.File) error {
		close(started)
		<-release
		return syncDurableFile(file)
	}
	done := make(chan error, 1)
	go func() { done <- h.EmitDurable(context.Background(), &message.Entry{Message: "confirmed"}) }()
	<-started
	select {
	case err := <-done:
		t.Fatalf("returned before sync: %v", err)
	default:
	}
	unblock()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(h.Path())
	if err != nil || string(content) != "confirmed\n" {
		t.Fatalf("file = %q, error = %v", content, err)
	}
}

func TestEmitDurableReportsWriteAndSyncFailures(t *testing.T) {
	for _, stage := range []string{"write", "sync", "sync-directory"} {
		t.Run(stage, func(t *testing.T) {
			h := newDurableTestHandler(t)
			failure := error(syscall.EIO)
			switch stage {
			case "write":
				failure = syscall.ENOSPC
				h.writeFile = func(file *os.File, buf []byte) (int, error) {
					n, err := file.Write(buf[:2])
					if err != nil {
						return n, err
					}
					return n, syscall.ENOSPC
				}
			case "sync":
				h.syncFile = func(*os.File) error { return failure }
			case "sync-directory":
				h.syncDir = func(string) error { return failure }
			}
			err := h.EmitDurable(context.Background(), &message.Entry{Message: "failure"})
			var operation *FileOperationError
			if !errors.Is(err, failure) || !errors.As(err, &operation) || operation.Stage != stage || operation.Path != h.Path() {
				t.Fatalf("error = %v, operation = %+v", err, operation)
			}
			// Do not continue appending after a partial write or failed sync.
			if err := h.EmitDurable(context.Background(), &message.Entry{Message: "later"}); !errors.Is(err, failure) {
				t.Fatalf("failed writer returned success on a later record: %v", err)
			}
		})
	}
}

func TestWriteAllHandlesShortWritesAndZeroProgress(t *testing.T) {
	for _, withError := range []bool{false, true} {
		var got []byte
		n, err := writeAll(nil, []byte("abcdef"), func(_ *os.File, buf []byte) (int, error) {
			length := len(buf)
			if length > 2 {
				length = 2
			}
			got = append(got, buf[:length]...)
			if withError && length < len(buf) {
				return length, io.ErrShortWrite
			}
			return length, nil
		})
		if n != 6 || err != nil || string(got) != "abcdef" {
			t.Fatalf("short writes: n=%d error=%v data=%q", n, err, got)
		}
	}
	n, err := writeAll(nil, []byte("x"), func(*os.File, []byte) (int, error) { return 0, nil })
	if n != 0 || !errors.Is(err, io.ErrNoProgress) {
		t.Fatalf("zero progress: n=%d error=%v", n, err)
	}
}

func TestEmitDurableCancellationDoesNotBlockWriter(t *testing.T) {
	h := newDurableTestHandler(t)
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	h.syncFile = func(file *os.File) error { close(started); <-release; return syncDurableFile(file) }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- h.EmitDurable(ctx, &message.Entry{Message: "accepted"}) }()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, ErrDurabilityUnknown) || !errors.Is(err, context.Canceled) {
		t.Fatalf("accepted cancellation = %v", err)
	}
	unblock()
	closed := make(chan error, 1)
	go func() { closed <- h.Close() }()
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("abandoned acknowledgement blocked Close")
	}
	if err := h.EmitDurable(context.Background(), &message.Entry{Message: "closed"}); !errors.Is(err, ErrHandlerClosed) {
		t.Fatalf("closed handler accepted a reliable request: %v", err)
	}
}

func TestEmitDurableCloseReleasesAdmissionAndDrains(t *testing.T) {
	cfg := newRegressionFileConfig(t)
	cfg.BufferSize = 1
	cfg.Backpressure = BackpressureConfig{Strategy: BackpressureStrategyDrop}
	h, err := NewFileHandler(cfg, formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"}), nil)
	if err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	var first, released sync.Once
	unblock := func() { released.Do(func() { close(release) }) }
	t.Cleanup(func() { unblock(); _ = h.Close() })
	h.syncFile = func(file *os.File) error {
		first.Do(func() { close(started); <-release })
		return syncDurableFile(file)
	}
	one, two, three := make(chan error, 1), make(chan error, 1), make(chan error, 1)
	go func() { one <- h.EmitDurable(context.Background(), &message.Entry{Message: "one"}) }()
	<-started
	go func() { two <- h.EmitDurable(context.Background(), &message.Entry{Message: "two"}) }()
	deadline := time.Now().Add(5 * time.Second)
	for len(h.entryChan) != 1 {
		if time.Now().After(deadline) {
			t.Fatal("second request did not enter file queue")
		}
		runtime.Gosched()
	}
	go func() { three <- h.EmitDurable(context.Background(), &message.Entry{Message: "three"}) }()
	closed := make(chan error, 1)
	go func() { closed <- h.Close() }()
	if err := <-three; !errors.Is(err, ErrHandlerClosed) {
		t.Fatalf("blocked admission = %v", err)
	}
	unblock()
	for _, done := range []chan error{one, two, closed} {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	content, err := os.ReadFile(h.Path())
	if err != nil || string(content) != "one\ntwo\n" {
		t.Fatalf("drained data = %q, error = %v", content, err)
	}
}

func TestOrdinaryBatchStopsAtDurableRequest(t *testing.T) {
	h := newDurableTestHandler(t)
	// Fill the queue directly before asking the sole writer to process it.
	// The formatter blocks its first ordinary record to make the batch boundary
	// deterministic while the reliable and following ordinary requests queue.
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	h.formatter = &callerBlockingFormatter{started: started, release: release}
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	if err := h.Emit(&message.Entry{Message: "before"}); err != nil {
		t.Fatal(err)
	}
	<-started
	done := make(chan error, 1)
	h.entryChan <- fileRequest{entry: message.Entry{Message: "durable"}, done: done}
	if err := h.Emit(&message.Entry{Message: "after"}); err != nil {
		t.Fatal(err)
	}
	var synced []string
	h.syncFile = func(file *os.File) error {
		data, err := os.ReadFile(file.Name())
		if err != nil {
			return err
		}
		synced = append(synced, string(data))
		return syncDurableFile(file)
	}
	unblock()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(synced, []string{"before\ndurable\n"}) {
		t.Fatalf("durable batch included a later ordinary record: %q", synced)
	}
}

func TestSizeRotationSynchronizesEachRename(t *testing.T) {
	cfg := newRegressionFileConfig(t)
	cfg.BackupCount = 3
	r, err := NewSizeRotator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	file, _, err := r.NeedRollover([]byte("active"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("active")); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{".1", ".2", ".3"} {
		if err := os.WriteFile(r.filePath+suffix, []byte(suffix), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var events []string
	r.fs.renameFile = func(old, next string) error {
		events = append(events, "rename:"+filepath.Base(old))
		return os.Rename(old, next)
	}
	r.fs.syncDir = func(path string) error {
		events = append(events, "directory")
		return syncDirectory(path)
	}
	r.fs.removeFile = func(path string) error {
		if path != r.intentPath() {
			t.Fatal("rotation unlinked a backup before renaming")
		}
		return os.Remove(path)
	}
	if _, err := r.DoRollover(); err != nil {
		t.Fatal(err)
	}
	var moves []string
	for index, event := range events {
		if strings.HasPrefix(event, "rename:") {
			if index+1 >= len(events) || events[index+1] != "directory" {
				t.Fatalf("rename was not followed by directory sync: %v", events)
			}
			if !strings.HasSuffix(event, ".glog-rotation.tmp") {
				moves = append(moves, event)
			}
		}
	}
	want := []string{"rename:app.log.2", "rename:app.log.1", "rename:app.log"}
	if !reflect.DeepEqual(moves, want) {
		t.Fatalf("rotation order = %v, want %v", moves, want)
	}
	content, _ := os.ReadFile(r.filePath + ".1")
	if string(content) != "active" {
		t.Fatalf("active file lost during rotation: %q", content)
	}
}

func TestRotationFailuresAreReportedAndNotRetried(t *testing.T) {
	for _, stage := range []string{"rename", "sync-directory"} {
		t.Run(stage, func(t *testing.T) {
			cfg := newRegressionFileConfig(t)
			cfg.BackupCount = 3
			r, err := NewSizeRotator(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			file, _, err := r.NeedRollover([]byte("confirmed"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := file.Write([]byte("confirmed")); err != nil {
				t.Fatal(err)
			}
			if err := syncDurableFile(file); err != nil {
				t.Fatal(err)
			}
			calls := 0
			if stage == "rename" {
				r.fs.renameFile = func(string, string) error { calls++; return syscall.EIO }
			} else {
				r.fs.syncDir = func(string) error { calls++; return syscall.EIO }
			}
			_, err = r.DoRollover()
			var operation *FileOperationError
			if !errors.Is(err, syscall.EIO) || !errors.As(err, &operation) || operation.Stage != stage {
				t.Fatalf("rotation error = %v", err)
			}
			if _, _, err := r.NeedRollover([]byte("later")); !errors.Is(err, syscall.EIO) {
				t.Fatalf("failed rotation was reused: %v", err)
			}
			if _, err := r.DoRollover(); !errors.Is(err, syscall.EIO) || calls != 1 {
				t.Fatalf("partial rotation retried: calls=%d error=%v", calls, err)
			}
		})
	}
}

func TestNestedDirectoryCreationPersistsParents(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "first", "second")
	fs := newRotationFS()
	var parents []string
	fs.syncDir = func(path string) error { parents = append(parents, path); return syncDirectory(path) }
	if err := fs.mkdir(path); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(parents, "\n")
	for _, parent := range []string{root, filepath.Join(root, "first")} {
		if !strings.Contains(joined, parent) {
			t.Fatalf("parent %s was not synced: %v", parent, parents)
		}
	}
	fs = newRotationFS()
	fs.syncDir = func(string) error { return syscall.EIO }
	if err := fs.mkdir(filepath.Join(root, "failure")); !errors.Is(err, syscall.EIO) {
		t.Fatalf("directory synchronization failure = %v", err)
	}
}

func TestDurableRecordsSurviveOrdinaryRotation(t *testing.T) {
	for name, kind := range map[string]RotatorType{
		"size": FileRotatorTypeSize, "time": FileRotatorTypeTime, "time-and-size": FileRotatorTypeTimeAndSize,
	} {
		t.Run(name, func(t *testing.T) {
			cfg := newRegressionFileConfig(t)
			cfg.RotatorType = kind
			cfg.BackupCount = 16
			cfg.Interval = 3600
			cfg.TimeSuffixFmt = "fixed"
			cfg.ReMatch = "^fixed$"
			if kind == FileRotatorTypeSize {
				cfg.MaxFileSize = 24
			}
			h, err := NewFileHandler(cfg, formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"}), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer h.Close()
			if err := h.EmitDurable(context.Background(), &message.Entry{Message: "confirmed-before"}); err != nil {
				t.Fatal(err)
			}
			switch r := h.rotator.(type) {
			case *TimeRotator:
				r.rolloverAt = 0
			case *TimeAndSizeRotator:
				r.rolloverAt = 0
			}
			if err := h.Emit(&message.Entry{Message: "ordinary"}); err != nil {
				t.Fatal(err)
			}
			if err := h.EmitDurable(context.Background(), &message.Entry{Message: "confirmed-after"}); err != nil {
				t.Fatal(err)
			}
			if err := h.Close(); err != nil {
				t.Fatal(err)
			}
			paths, err := filepath.Glob(filepath.Join(cfg.FileDir, "app.log*"))
			if err != nil {
				t.Fatal(err)
			}
			var content strings.Builder
			for _, path := range paths {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				content.Write(data)
			}
			for _, message := range []string{"confirmed-before", "ordinary", "confirmed-after"} {
				if strings.Count(content.String(), message+"\n") != 1 {
					t.Fatalf("record %q missing/duplicated across rotation: %q", message, content.String())
				}
			}
		})
	}
}
