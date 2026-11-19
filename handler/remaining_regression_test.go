package handler

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

type failingFormatter struct{}

func (failingFormatter) Format(*message.Entry) ([]byte, error) {
	return nil, errors.New("format failed")
}

func TestErrorCallbackCanCloseFileHandler(t *testing.T) {
	callbackDone := make(chan error, 1)
	var fileHandler *FileHandler
	var err error
	cfg := newRegressionFileConfig(t)
	cfg.ErrCallback = func(interface{}, error) { callbackDone <- fileHandler.Close() }
	fileHandler, err = NewFileHandler(cfg, failingFormatter{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := fileHandler.Emit(&message.Entry{Message: "trigger"}); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-callbackDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("error callback deadlocked closing its file handler")
	}
}

func TestRotatorsRejectOversizedBatchesAndAcceptExactLimit(t *testing.T) {
	for _, kind := range []RotatorType{FileRotatorTypeSize, FileRotatorTypeTimeAndSize} {
		cfg := newRegressionFileConfig(t)
		cfg.RotatorType = kind
		cfg.MaxFileSize = 16
		cfg.Interval = 3600
		rotator, err := NewRotator(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rotator.Close() }()
		if _, rotate, err := rotator.NeedRollover(make([]byte, 17)); !errors.Is(err, ErrFileSizeLimit) || rotate {
			t.Fatalf("rotator %d oversized batch: rotate=%t error=%v", kind, rotate, err)
		}
		fileHandler := &FileHandler{rotator: rotator}
		if err := fileHandler.realWrite(make([]byte, 16)); err != nil {
			t.Fatalf("rotator %d exact limit: %v", kind, err)
		}
		info, err := os.Stat(filepath.Join(cfg.FileDir, "app.log"))
		if err != nil || info.Size() != 16 {
			t.Fatalf("rotator %d exact limit: info=%v error=%v", kind, info, err)
		}
	}
}

func TestExpiredFileDoesNotBypassSizeLimit(t *testing.T) {
	cfg := newRegressionFileConfig(t)
	cfg.RotatorType = FileRotatorTypeTimeAndSize
	cfg.MaxFileSize = 16
	cfg.Interval = 3600
	path := filepath.Join(cfg.FileDir, "app.log")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	old := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	errorsReported := make(chan error, 1)
	cfg.ErrCallback = func(_ interface{}, err error) { errorsReported <- err }
	fm := formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"})
	fileHandler, err := NewFileHandler(cfg, fm, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := fileHandler.Emit(&message.Entry{Message: strings.Repeat("x", 64)}); err != nil {
		t.Fatal(err)
	}
	if err := fileHandler.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-errorsReported:
		if !errors.Is(err, ErrFileSizeLimit) {
			t.Fatalf("callback error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("oversized batch did not report its error")
	}
	info, err := os.Stat(path)
	if err != nil || info.Size() > cfg.MaxFileSize {
		t.Fatalf("file exceeded limit: info=%v error=%v", info, err)
	}
}

func TestTimeRolloverRechecksExistingConcurrentFileQuota(t *testing.T) {
	cfg := newRegressionFileConfig(t)
	cfg.RotatorType = FileRotatorTypeTimeAndSize
	cfg.MaxFileSize = 16
	cfg.Interval = 3600
	cfg.ConcurrentlyWrite = true
	cfg.TimeSuffixFmt = "fixed"
	rotator, err := NewTimeAndSizeRotator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rotator.Close() }()
	if err := os.WriteFile(rotator.filePath, []byte(strings.Repeat("x", 15)), 0600); err != nil {
		t.Fatal(err)
	}
	rotator.rolloverAt = time.Now().Unix() - 1
	fileHandler := &FileHandler{rotator: rotator}
	if err := fileHandler.realWrite([]byte("xx")); !errors.Is(err, ErrFileSizeLimit) {
		t.Fatalf("rollover quota error = %v", err)
	}
	info, err := os.Stat(rotator.filePath)
	if err != nil || info.Size() != 15 {
		t.Fatalf("rollover changed over-quota file: info=%v error=%v", info, err)
	}
}
