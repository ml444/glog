package handler

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
)

type redactingFormatter struct {
	*formatter.TextFormatter
}

func (fm *redactingFormatter) Format(*message.Entry) ([]byte, error) {
	return []byte("[REDACTED]\n"), nil
}

func newRegressionFileConfig(t *testing.T) *FileHandlerConfig {
	t.Helper()
	return &FileHandlerConfig{
		FileDir:       t.TempDir(),
		FileName:      "app",
		FileSuffix:    "log",
		MaxFileSize:   1 << 20,
		BufferSize:    8,
		BulkWriteSize: 256,
		RotatorType:   FileRotatorTypeSize,
		Backpressure:  BackpressureConfig{Strategy: BackpressureStrategyBlock},
	}
}

func TestFileHandlerRespectsEmbeddedFormatterOverride(t *testing.T) {
	cfg := newRegressionFileConfig(t)
	fm := &redactingFormatter{
		TextFormatter: formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"}),
	}
	fileHandler, err := NewFileHandler(cfg, fm, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = fileHandler.Close() })
	if err := fileHandler.Emit(&message.Entry{Message: "SECRET", Level: level.InfoLevel}); err != nil {
		t.Fatal(err)
	}
	if err := fileHandler.Close(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(cfg.FileDir, "app.log"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "[REDACTED]\n" {
		t.Fatalf("custom Format was bypassed: %q", content)
	}
}

type callerBlockingFormatter struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (fm *callerBlockingFormatter) Format(entry *message.Entry) ([]byte, error) {
	fm.once.Do(func() {
		close(fm.started)
		<-fm.release
	})
	if entry.Caller != nil {
		return []byte(entry.Message + " " + entry.Caller.File + "\n"), nil
	}
	return []byte(entry.Message + "\n"), nil
}

func TestFileHandlerSnapshotsEntryAndCaller(t *testing.T) {
	cfg := newRegressionFileConfig(t)
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
	entry := &message.Entry{Message: "before", Caller: &runtime.Frame{File: "before.go"}}
	if err := fileHandler.Emit(entry); err != nil {
		t.Fatal(err)
	}
	entry.Message = "after"
	entry.Caller.File = "after.go"
	release()
	if err := fileHandler.Close(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(cfg.FileDir, "app.log"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "first\nbefore before.go\n" {
		t.Fatalf("Entry changed after Emit returned: %q", content)
	}
}

func TestFileHandlerErrorCallbackOwnsBuffer(t *testing.T) {
	cfg := newRegressionFileConfig(t)
	cfg.RotatorType = FileRotatorTypeTimeAndSize
	cfg.MaxFileSize = 1
	cfg.Interval = 3600
	cfg.TimeSuffixFmt = "2006010215"
	cfg.ReMatch = `^\d{10}$`
	failedBatches := make(chan []byte, 2)
	cfg.ErrCallback = func(value interface{}, _ error) { failedBatches <- value.([]byte) }
	fm := formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"})
	fileHandler, err := NewFileHandler(cfg, fm, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = fileHandler.Close() })
	firstMessage := strings.Repeat("first", 32)
	secondMessage := strings.Repeat("later", 32)
	if err := fileHandler.Emit(&message.Entry{Message: firstMessage}); err != nil {
		t.Fatal(err)
	}
	var saved []byte
	select {
	case saved = <-failedBatches:
	case <-time.After(time.Second):
		t.Fatal("missing first error callback")
	}
	if err := fileHandler.Emit(&message.Entry{Message: secondMessage}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-failedBatches:
	case <-time.After(time.Second):
		t.Fatal("missing second error callback")
	}
	if err := fileHandler.Close(); err != nil {
		t.Fatal(err)
	}
	if string(saved) != firstMessage+"\n" {
		t.Fatalf("saved error buffer overwritten: %q", saved)
	}
}
