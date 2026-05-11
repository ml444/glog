//go:build !windows && !plan9

package log

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// envSignalFlushChild enables the child side of TestSignalTermFlushesFileUnderLoad.
const envSignalFlushChild = "GLOG_SIGNAL_FLUSH_CHILD"

// TestSignalTermFlushesFileUnderLoad spawns a subprocess that:
// 1) Stops the package-default logger, then InitLog with a JSON file worker (small buffers).
// 2) Starts many goroutines that continuously Infof to stress the async pipeline.
// 3) Blocks until SIGTERM; glog's signal handler calls Stop() then os.Exit(1).
//
// The parent sends SIGTERM after a load window, then asserts the log file contains
// a large number of unique markers (i.e. data reached the file before exit).
func TestSignalTermFlushesFileUnderLoad(t *testing.T) {
	if os.Getenv(envSignalFlushChild) == "1" {
		runSignalTermFlushesFileChild(t)
		return
	}

	dir := t.TempDir()
	t.Logf("dir: %s", dir)
	cmd := exec.Command(os.Args[0], "-test.run=TestSignalTermFlushesFileUnderLoad", "-test.count=1", "-test.timeout=90s")
	cmd.Env = append(os.Environ(),
		envSignalFlushChild+"=1",
		"GLOG_TEST_LOG_DIR="+dir,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	pidFile := filepath.Join(dir, "child.pid")
	deadline := time.Now().Add(15 * time.Second)
	var pid int
	for time.Now().Before(deadline) {
		if b, err := os.ReadFile(pidFile); err == nil {
			pid, _ = strconv.Atoi(strings.TrimSpace(string(b)))
			if pid > 0 {
				break
			}
		}
		time.Sleep(15 * time.Millisecond)
	}
	if pid == 0 {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("child pid file not ready in time; output:\n%s", out.String())
	}

	// Let the child build up concurrent backlog in channels / file buffers.
	time.Sleep(450 * time.Millisecond)

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("SIGTERM: %v", err)
	}

	waitErr := cmd.Wait()
	if cmd.ProcessState == nil {
		t.Fatal("missing process state after Wait")
	}
	if ec := cmd.ProcessState.ExitCode(); ec != 1 {
		t.Fatalf("child exit code %d want 1 (signal path uses os.Exit(1)); wait err=%v; output:\n%s", ec, waitErr, out.String())
	}

	logPath := filepath.Join(dir, "sigchild.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file %s: %v", logPath, err)
	}
	n := bytes.Count(data, []byte("SIGLINE:"))
	// Under heavy load, expect a large number of markers persisted after Stop/close.
	const minLines = 2000
	if n < minLines {
		t.Fatalf("expected at least %d flushed markers in file, got %d (file bytes %d)", minLines, n, len(data))
	}
}

func runSignalTermFlushesFileChild(t *testing.T) {
	dir := os.Getenv("GLOG_TEST_LOG_DIR")
	if dir == "" {
		t.Fatal("GLOG_TEST_LOG_DIR not set")
	}

	// Drop the logger created in package init so InitLog can own the only engine.
	Stop()

	jsonCfg := NewDefaultJSONFormatterConfig()
	jsonCfg.PrettyPrint = false

	err := InitLog(
		SetLoggerName("sigchild"),
		SetLoggerLevel(DebugLevel),
		SetWorkerConfigs(
			NewWorkerConfig(PrintLevel, 96).SetFileHandlerConfig(
				NewDefaultFileHandlerConfig(dir).
					WithFileName("sigchild").
					WithRotatorType(FileRotatorTypeSize).
					WithFileSize(1 << 30).
					WithBulkSize(256).
					WithBufferSize(3000),
			).SetJSONFormatterConfig(jsonCfg),
		),
	)
	if err != nil {
		t.Fatalf("InitLog: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "child.pid"), []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatalf("write pid: %v", err)
	}

	const workers = 128
	var seq int64
	for i := 0; i < workers; i++ {
		go func() {
			for {
				Infof("SIGLINE:%d", atomic.AddInt64(&seq, 1))
				runtime.Gosched()
			}
		}()
	}

	// Block until SIGTERM triggers registerShutdownOnSignals → Stop → os.Exit(1).
	select {}
}
