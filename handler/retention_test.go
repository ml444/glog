package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestTimeRotatorRetainsOnlyExcessBackups(t *testing.T) {
	for _, existing := range []int{0, 2, 3} {
		t.Run(fmt.Sprintf("existing-%d", existing), func(t *testing.T) {
			cfg := newRegressionFileConfig(t)
			cfg.BackupCount = 3
			cfg.Interval = 5
			cfg.TimeSuffixFmt = "2006010215"
			cfg.ReMatch = `^\d{10}$`
			for index := 0; index < existing; index++ {
				name := filepath.Join(cfg.FileDir, fmt.Sprintf("app.log.200001010%d", index))
				if err := os.WriteFile(name, []byte("backup"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			unrelated := filepath.Join(cfg.FileDir, "other.log.2000010100")
			if err := os.WriteFile(unrelated, []byte("unrelated"), 0o600); err != nil {
				t.Fatal(err)
			}
			rotator, err := NewTimeRotator(cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = rotator.Close() })
			if rotator.interval != cfg.Interval {
				t.Fatalf("interval = %d, want %d", rotator.interval, cfg.Interval)
			}
			file, _, err := rotator.NeedRollover([]byte("current"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := file.Write([]byte("current")); err != nil {
				t.Fatal(err)
			}
			if _, err := rotator.DoRollover(); err != nil {
				t.Fatal(err)
			}
			backups, err := filepath.Glob(filepath.Join(cfg.FileDir, "app.log.*"))
			if err != nil {
				t.Fatal(err)
			}
			want := existing + 1
			if want > cfg.BackupCount {
				want = cfg.BackupCount
			}
			if len(backups) != want {
				t.Fatalf("backups = %v, want %d", backups, want)
			}
			if existing == 3 {
				if _, err := os.Stat(filepath.Join(cfg.FileDir, "app.log.2000010100")); !os.IsNotExist(err) {
					t.Fatalf("oldest backup was not removed: %v", err)
				}
			}
			if _, err := os.Stat(unrelated); err != nil {
				t.Fatalf("unrelated backup removed: %v", err)
			}
		})
	}
}

func TestTimeRotatorsDoNotOverwriteRepeatedSuffix(t *testing.T) {
	for _, kind := range []RotatorType{FileRotatorTypeTime, FileRotatorTypeTimeAndSize} {
		t.Run(fmt.Sprintf("rotator-%d", kind), func(t *testing.T) {
			cfg := newRegressionFileConfig(t)
			cfg.RotatorType = kind
			cfg.Interval = 3600
			cfg.BackupCount = 2
			cfg.TimeSuffixFmt = "fixed"
			cfg.ReMatch = `^fixed$`
			rotator, err := NewRotator(cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = rotator.Close() })
			for round := 0; round < 4; round++ {
				payload := []byte(fmt.Sprintf("round-%d", round))
				file, _, err := rotator.NeedRollover(payload)
				if err != nil {
					t.Fatal(err)
				}
				written, err := file.Write(payload)
				if err != nil {
					t.Fatal(err)
				}
				rotator.RecordBytesWritten(written)
				if _, err := rotator.DoRollover(); err != nil {
					t.Fatal(err)
				}
			}
			paths, err := filepath.Glob(filepath.Join(cfg.FileDir, "app.log.fixed*"))
			if err != nil {
				t.Fatal(err)
			}
			if len(paths) != 2 {
				t.Fatalf("backups = %v, want 2", paths)
			}
			sort.Strings(paths)
			for index, name := range paths {
				content, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if want := fmt.Sprintf("round-%d", index+2); string(content) != want {
					t.Fatalf("backup %q content = %q, want %q", name, content, want)
				}
			}
		})
	}
}

func TestTimeAndSizeRotatorMatchesActualBackupNames(t *testing.T) {
	for _, concurrent := range []bool{false, true} {
		t.Run(fmt.Sprintf("concurrent-%t", concurrent), func(t *testing.T) {
			cfg := newRegressionFileConfig(t)
			cfg.BackupCount = 2
			cfg.Interval = 3600
			cfg.TimeSuffixFmt = "2006010215"
			cfg.ReMatch = `^\d{10}(\.\w+)?$`
			cfg.ConcurrentlyWrite = concurrent
			for index := 0; index < 3; index++ {
				name := fmt.Sprintf("app.log.200001010%d", index)
				if concurrent {
					name = fmt.Sprintf("app.200001010%d.log", index)
				}
				if err := os.WriteFile(filepath.Join(cfg.FileDir, name), []byte("backup"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			rotator, err := NewTimeAndSizeRotator(cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = rotator.Close() })
			file, _, err := rotator.NeedRollover([]byte("current"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := file.Write([]byte("current")); err != nil {
				t.Fatal(err)
			}
			if _, err := rotator.DoRollover(); err != nil {
				t.Fatal(err)
			}
			pattern := "app.log.*"
			if concurrent {
				pattern = "app.*.log"
			}
			backups, err := filepath.Glob(filepath.Join(cfg.FileDir, pattern))
			if err != nil {
				t.Fatal(err)
			}
			if len(backups) != cfg.BackupCount {
				t.Fatalf("backups = %v, want %d", backups, cfg.BackupCount)
			}
		})
	}
}
