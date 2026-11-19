//go:build linux || darwin

package log

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const durableCrashChild = "GLOG_DURABLE_CRASH_CHILD"

func TestWriteDurableSurvivesAbruptExit(t *testing.T) {
	if mode := os.Getenv(durableCrashChild); mode != "" {
		runDurableCrashChild(t, mode)
		return
	}
	for _, mode := range []string{"exit", "panic", "kill"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWriteDurableSurvivesAbruptExit$", "-test.count=1")
			cmd.Env = append(os.Environ(), durableCrashChild+"="+mode, "GLOG_DURABLE_CRASH_DIR="+dir, EnvNoSignalShutdown+"=1")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			var confirmed []string
			scanner := bufio.NewScanner(stdout)
			killed := false
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "ACK:") {
					confirmed = append(confirmed, strings.TrimPrefix(line, "ACK:"))
					if mode == "kill" && len(confirmed) >= 8 && !killed {
						killed = true
						if err := cmd.Process.Kill(); err != nil {
							t.Errorf("Kill: %v", err)
						}
					}
				}
			}
			scanErr := scanner.Err()
			waitErr := cmd.Wait()
			if scanErr != nil || ctx.Err() != nil || waitErr == nil {
				t.Fatalf("child did not exit abruptly: scanner=%v context=%v wait=%v stderr=%s", scanErr, ctx.Err(), waitErr, stderr.String())
			}
			if len(confirmed) < 8 || mode != "kill" && len(confirmed) != 12 {
				t.Fatalf("unexpected acknowledged IDs: %v; stderr=%s", confirmed, stderr.String())
			}
			paths, err := filepath.Glob(filepath.Join(dir, "crash.log*"))
			if err != nil {
				t.Fatal(err)
			}
			var content []byte
			for _, path := range paths {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				content = append(content, data...)
			}
			for _, id := range confirmed {
				if count := bytes.Count(content, []byte(id+"\n")); count != 1 {
					t.Fatalf("confirmed ID %q appears %d times after %s; files=%v", id, count, mode, paths)
				}
			}
		})
	}
}

func runDurableCrashChild(t *testing.T, mode string) {
	logger, err := New(SetWorkerConfigs(NewWorkerConfig(InfoLevel, 2).
		SetFileHandlerConfig(NewDefaultFileHandlerConfig(os.Getenv("GLOG_DURABLE_CRASH_DIR")).
			WithFileName("crash").WithRotatorType(FileRotatorTypeSize).
			WithFileSize(48).WithBackupCount(64).WithBufferSize(2)).
		SetTextFormatterConfig(&TextFormatterConfig{PatternStyle: "%[Message]v"})))
	if err != nil {
		t.Fatal(err)
	}
	// Deliberately do not call Stop, including from a defer or panic handler.
	for index := 0; index < 12 || mode == "kill"; index++ {
		id := fmt.Sprintf("DURABLE:%06d", index)
		if err := logger.WriteDurable(context.Background(), InfoLevel, id); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintln(os.Stdout, "ACK:"+id)
		// Exercise ordinary writes triggering rotations after confirmed records.
		logger.Info("unconfirmed tail")
	}
	if mode == "panic" {
		panic("intentional crash after durable acknowledgements")
	}
	os.Exit(23)
}
