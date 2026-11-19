//go:build linux || darwin

package handler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ml444/glog/formatter"
)

func TestSizeRotationRecoversAtEveryPersistenceBoundary(t *testing.T) {
	run := func(t *testing.T, cut int) int {
		cfg := newRegressionFileConfig(t)
		cfg.BackupCount = 3
		cfg.MaxFileSize = 32
		r, err := NewSizeRotator(cfg)
		if err != nil {
			t.Fatal(err)
		}
		for index, record := range []string{"ACTIVE-D\n", "BACKUP-C\n", "BACKUP-B\n", "EXPIRED-A\n"} {
			path := r.filePath
			if index > 0 {
				path += fmt.Sprintf(".%d", index)
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
			if err != nil {
				t.Fatal(err)
			}
			_, writeErr := file.WriteString(record)
			syncErr := syncDurableFile(file)
			closeErr := file.Close()
			if writeErr != nil || syncErr != nil || closeErr != nil {
				t.Fatalf("seed failed: %v %v %v", writeErr, syncErr, closeErr)
			}
		}
		if _, _, err := r.NeedRollover([]byte("next")); err != nil {
			t.Fatal(err)
		}
		steps := 0
		fault := errors.New("simulated interruption after persistence operation")
		after := func(err error) error {
			steps++
			if err == nil && steps == cut {
				return fault
			}
			return err
		}
		syncFile, syncDir := r.fs.syncFile, r.fs.syncDir
		rename, remove := r.fs.renameFile, r.fs.removeFile
		r.fs.syncFile = func(file *os.File) error { return after(syncFile(file)) }
		r.fs.syncDir = func(path string) error { return after(syncDir(path)) }
		r.fs.renameFile = func(from, to string) error { return after(rename(from, to)) }
		r.fs.removeFile = func(path string) error { return after(remove(path)) }
		_, err = r.DoRollover()
		rotationSteps := steps
		if cut > 0 && !errors.Is(err, fault) {
			t.Fatalf("cut %d did not interrupt rotation: %v", cut, err)
		}
		_ = r.Close()
		// A fresh instance models restart: no in-memory progress survives.
		r, err = NewSizeRotator(cfg)
		if err != nil {
			t.Fatalf("recovery at cut %d: %v", cut, err)
		}
		defer r.Close()
		payload := []byte("NEXT-RECORD-IS-LONG-ENOUGH\n")
		file, rotate, err := r.NeedRollover(payload)
		if err != nil {
			t.Fatal(err)
		}
		if rotate {
			file, err = r.DoRollover()
			if err != nil {
				t.Fatal(err)
			}
		}
		if _, err := file.Write(payload); err != nil {
			t.Fatal(err)
		}
		if err := syncDurableFile(file); err != nil {
			t.Fatal(err)
		}
		paths, _ := filepath.Glob(r.filePath + "*")
		var content strings.Builder
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			content.Write(data)
		}
		for _, id := range []string{"ACTIVE-D", "BACKUP-C", "BACKUP-B", "NEXT-RECORD-IS-LONG-ENOUGH"} {
			if strings.Count(content.String(), id+"\n") != 1 {
				t.Fatalf("cut %d lost/duplicated %s: %q", cut, id, content.String())
			}
		}
		if _, err := os.Stat(r.intentPath()); !os.IsNotExist(err) {
			t.Fatalf("completed recovery left a rotation journal: %v", err)
		}
		return rotationSteps
	}
	steps := run(t, 0)
	for cut := 1; cut <= steps; cut++ {
		t.Run(fmt.Sprintf("cut-%d", cut), func(t *testing.T) { run(t, cut) })
	}
}

func TestPendingSizeRotationRejectsTypeChange(t *testing.T) {
	cfg := newRegressionFileConfig(t)
	cfg.BackupCount = 3
	r, err := NewSizeRotator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	file, _, err := r.NeedRollover([]byte("confirmed\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("confirmed\n"); err != nil {
		t.Fatal(err)
	}
	if err := r.fs.closeFile(&r.file); err != nil {
		t.Fatal(err)
	}
	if err := r.saveIntent(sizeRotationIntent{Version: 1, Next: 3, BackupCount: 3}); err != nil {
		t.Fatal(err)
	}
	changed := *cfg
	changed.RotatorType = FileRotatorTypeTime
	fm := formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"})
	h, err := NewFileHandler(&changed, fm, nil)
	if h != nil {
		_ = h.Close()
	}
	var operation *FileOperationError
	if !errors.As(err, &operation) || operation.Stage != "recover-rotation" {
		t.Fatalf("unfinished rotation allowed a type change: %v", err)
	}
	r, err = NewSizeRotator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(r.filePath + ".1")
	if err != nil || string(data) != "confirmed\n" {
		t.Fatalf("recovery lost the confirmed file: %q %v", data, err)
	}
	h, err = NewFileHandler(&changed, fm, nil)
	if err != nil {
		t.Fatalf("type change still blocked after recovery: %v", err)
	}
	_ = h.Close()
}
