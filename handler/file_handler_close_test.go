package handler

import (
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
)

type blockingFormatter struct {
	started chan struct{}
	release chan struct{}
	count   uint64
}

func (f *blockingFormatter) Format(entry *message.Entry) ([]byte, error) {
	if atomic.AddUint64(&f.count, 1) == 1 {
		close(f.started)
		<-f.release
	}
	return []byte(entry.Message + "\n"), nil
}

func TestFileHandlerCloseTwiceDoesNotHang(t *testing.T) {
	dir := t.TempDir()
	cfg := &FileHandlerConfig{
		FileDir:       dir,
		FileName:      "c",
		FileSuffix:    "log",
		MaxFileSize:   1 << 20,
		BufferSize:    8,
		BulkWriteSize: 256,
		RotatorType:   FileRotatorTypeSize,
	}
	fm := formatter.NewTextFormatter(formatter.TextFormatterConfig{
		BaseFormatterConfig: formatter.BaseFormatterConfig{
			LoggerName: "t",
			TimeLayout: "2006-01-02 15:04:05",
		},
		PatternStyle: "%[Message]v",
	})
	h, err := NewFileHandler(cfg, fm, nil)
	if err != nil {
		t.Fatal(err)
	}
	entry := &message.Entry{
		Message: "hello",
		Level:   level.InfoLevel,
		Time:    time.Now(),
	}
	if err := h.Emit(entry); err != nil {
		t.Fatal(err)
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestFileHandlerDropsBeforeFormatting(t *testing.T) {
	dir := t.TempDir()
	cfg := &FileHandlerConfig{
		FileDir:       dir,
		FileName:      "deferred",
		FileSuffix:    "log",
		MaxFileSize:   1 << 20,
		BufferSize:    1,
		BulkWriteSize: 256,
		RotatorType:   FileRotatorTypeSize,
		Backpressure:  BackpressureConfig{Strategy: BackpressureStrategyDrop},
	}
	fm := &blockingFormatter{started: make(chan struct{}), release: make(chan struct{})}
	h, err := NewFileHandler(cfg, fm, nil)
	if err != nil {
		t.Fatal(err)
	}

	entry := func(messageText string) *message.Entry {
		return &message.Entry{Message: messageText, Level: level.InfoLevel, Time: time.Now()}
	}
	if err := h.Emit(entry("first")); err != nil {
		t.Fatal(err)
	}
	<-fm.started
	if err := h.Emit(entry("second")); err != nil {
		t.Fatal(err)
	}
	if err := h.Emit(entry("dropped")); !errors.Is(err, ErrBackpressureDropped) {
		t.Fatalf("third Emit error = %v, want %v", err, ErrBackpressureDropped)
	}
	if got := atomic.LoadUint64(&fm.count); got != 1 {
		t.Fatalf("formatter calls before release = %d, want 1", got)
	}

	close(fm.release)
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadUint64(&fm.count); got != 2 {
		t.Fatalf("formatter calls after close = %d, want 2", got)
	}
	content, err := os.ReadFile(filepath.Join(dir, "deferred.log"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(content), "first\nsecond\n"; got != want {
		t.Fatalf("file content = %q, want %q", got, want)
	}
}

func TestFileHandlerCloseWithoutEntries(t *testing.T) {
	cfg := &FileHandlerConfig{
		FileDir:       t.TempDir(),
		FileName:      "empty",
		FileSuffix:    "log",
		MaxFileSize:   1 << 20,
		BufferSize:    1,
		BulkWriteSize: 256,
		RotatorType:   FileRotatorTypeSize,
	}
	fm := formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"})
	h, err := NewFileHandler(cfg, fm, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
}
