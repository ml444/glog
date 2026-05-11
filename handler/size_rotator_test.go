package handler

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestSizeRotatorTrackedSizeTriggersRollover(t *testing.T) {
	dir := t.TempDir()
	cfg := &FileHandlerConfig{
		FileDir:     dir,
		FileName:    "sz",
		FileSuffix:  "log",
		MaxFileSize: 200,
	}
	r, err := NewSizeRotator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()

	payload := make([]byte, 100)
	_, rotate, err := r.NeedRollover(payload)
	if err != nil {
		t.Fatal(err)
	}
	if rotate {
		t.Fatal("unexpected rotate on empty file")
	}
	r.RecordBytesWritten(100)
	_, rotate, err = r.NeedRollover(make([]byte, 101))
	if err != nil {
		t.Fatal(err)
	}
	if !rotate {
		t.Fatal("expected rotate when tracked size + next write exceeds max")
	}
}

func TestCollectRotatedBackupPaths(t *testing.T) {
	dir := t.TempDir()
	re := regexp.MustCompile(`^\d+$`)
	for _, name := range []string{"app.1", "app.2", "other.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	paths, err := collectRotatedBackupPaths(dir, "app", re)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Fatalf("paths = %v (%d), want 2", paths, len(paths))
	}
}
