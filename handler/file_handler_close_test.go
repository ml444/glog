package handler

import (
	"testing"
	"time"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
)

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
