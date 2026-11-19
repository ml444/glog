package log

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ml444/glog/message"
)

func requireDurablePlatform(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("strong file synchronization is only implemented on Linux and macOS")
	}
}

func TestWriteDurableUnsupportedPlatformPreservesOrdinaryLogging(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		t.Skip("this platform supports durable writes")
	}
	worker, path := durableTestWorker(t, "unsupported")
	logger, err := New(SetWorkerConfigs(worker))
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Stop()
	if err := logger.WriteDurable(context.Background(), InfoLevel, "unsupported"); !errors.Is(err, ErrDurableUnsupported) {
		t.Fatalf("unsupported platform = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("unsupported durable write opened a file: %v", err)
	}
	logger.Info("ordinary")
	if err := logger.Stop(); err != nil {
		t.Fatal(err)
	}
	if got := readDurableTestFile(t, path); got != "ordinary\n" {
		t.Fatalf("ordinary logging changed: %q", got)
	}
}

func durableTestWorker(t *testing.T, name string) (*WorkerConfig, string) {
	t.Helper()
	dir := t.TempDir()
	worker := NewWorkerConfig(InfoLevel, 1).
		SetFileHandlerConfig(NewDefaultFileHandlerConfig(dir).
			WithFileName(name).WithRotatorType(FileRotatorTypeSize).
			WithBufferSize(1).WithFileSize(1 << 20)).
		SetTextFormatterConfig(&TextFormatterConfig{PatternStyle: "%[Message]v"})
	return worker, filepath.Join(dir, name+".log")
}

func readDurableTestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestWriteDurableAllFilesBypassesDropAndSample(t *testing.T) {
	requireDurablePlatform(t)
	first, firstPath := durableTestWorker(t, "first")
	second, secondPath := durableTestWorker(t, "second")
	first.SetBackpressureStrategy(BackpressureStrategyDrop)
	first.HandlerCfg.File.WithBackpressureStrategy(BackpressureStrategyDrop)
	second.SetBackpressureStrategy(BackpressureStrategySample)
	second.HandlerCfg.File.WithBackpressureStrategy(BackpressureStrategySample)
	custom := &collectHandler{}
	var exited bool
	logger, err := New(SetWorkerConfigs(first, second, NewWorkerConfig(InfoLevel, 1).SetHandler(custom)), SetThrowOnLevel(FatalLevel), SetExitFunc(func(int) { exited = true }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = logger.Stop() })
	var wg sync.WaitGroup
	for g := 0; g < 6; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for n := 0; n < 8; n++ {
				if err := logger.WriteDurable(context.Background(), FatalLevel, fmt.Sprintf("id:%d:%d", g, n)); err != nil {
					t.Errorf("WriteDurable: %v", err)
				}
			}
		}(g)
	}
	wg.Wait()
	if exited {
		t.Fatal("WriteDurable triggered ExitFunc")
	}
	if custom.len() != 0 {
		t.Fatal("WriteDurable dispatched to a custom handler")
	}
	// Read before Stop: confirmation must come from the write, not shutdown.
	for _, path := range []string{firstPath, secondPath} {
		content := readDurableTestFile(t, path)
		if strings.Count(content, "id:") != 48 {
			t.Fatalf("%s: missing confirmed records: %q", path, content)
		}
		for g := 0; g < 6; g++ {
			for n := 0; n < 8; n++ {
				if strings.Count(content, fmt.Sprintf("id:%d:%d\n", g, n)) != 1 {
					t.Fatalf("%s: missing or duplicate id %d:%d", path, g, n)
				}
			}
		}
	}
}

type durableFilterFunc func(*message.Entry) bool

func (f durableFilterFunc) Filter(entry *message.Entry) bool { return f(entry) }

type durableFormatFunc func(*message.Entry) ([]byte, error)

func (f durableFormatFunc) Format(entry *message.Entry) ([]byte, error) { return f(entry) }

func TestWriteDurableCopiesEntriesPerFile(t *testing.T) {
	requireDurablePlatform(t)
	first, firstPath := durableTestWorker(t, "first")
	second, secondPath := durableTestWorker(t, "second")
	first.SetFilter(durableFilterFunc(func(entry *message.Entry) bool { entry.Message = "redacted"; return true }))
	logger, err := New(SetWorkerConfigs(first, second))
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Stop()
	if err := logger.WriteDurable(context.Background(), InfoLevel, "original"); err != nil {
		t.Fatal(err)
	}
	if got := readDurableTestFile(t, firstPath); got != "redacted\n" {
		t.Fatalf("first = %q", got)
	}
	if got := readDurableTestFile(t, secondPath); got != "original\n" {
		t.Fatalf("second = %q", got)
	}
}

func TestWriteDurableReportsAllFailuresAndPartialSuccess(t *testing.T) {
	requireDurablePlatform(t)
	good, goodPath := durableTestWorker(t, "good")
	bad, _ := durableTestWorker(t, "bad")
	filtered, _ := durableTestWorker(t, "filtered")
	formatErr := errors.New("format failed")
	bad.SetFormatter(durableFormatFunc(func(*message.Entry) ([]byte, error) { return nil, formatErr }))
	filtered.SetFilter(durableFilterFunc(func(*message.Entry) bool { return false }))
	logger, err := New(SetWorkerConfigs(good, bad, filtered))
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Stop()
	err = logger.WriteDurable(context.Background(), InfoLevel, "saved")
	var failures *DurableWriteError
	if !errors.As(err, &failures) || len(failures.Failures) != 2 {
		t.Fatalf("error = %v, want two target failures", err)
	}
	if !errors.Is(err, formatErr) || !errors.Is(err, ErrLogFiltered) {
		t.Fatalf("causes were lost: %v", err)
	}
	for _, failure := range failures.Failures {
		if failure.Path == "" || failure.Stage == "wait" {
			t.Fatalf("missing operation details: %+v", failure)
		}
	}
	if got := readDurableTestFile(t, goodPath); got != "saved\n" {
		t.Fatalf("successful target was not retained: %q", got)
	}
}

func TestWriteDurableRejectsMissingFilteredAndEmptyTargets(t *testing.T) {
	requireDurablePlatform(t)
	worker, _ := durableTestWorker(t, "empty")
	worker.SetFormatter(durableFormatFunc(func(*message.Entry) ([]byte, error) { return nil, nil }))
	logger, err := New(SetWorkerConfigs(worker), SetLoggerLevel(InfoLevel))
	if err != nil {
		t.Fatal(err)
	}
	if err := logger.WriteDurable(context.Background(), DebugLevel, "filtered"); !errors.Is(err, ErrLogFiltered) {
		t.Fatalf("filtered = %v", err)
	}
	if err := logger.WriteDurable(context.Background(), InfoLevel, "empty"); !errors.Is(err, ErrEmptyLogRecord) {
		t.Fatalf("empty = %v", err)
	}
	if err := logger.WriteDurable(context.Background(), NoneLevel, "invalid"); !errors.Is(err, ErrInvalidLevel) {
		t.Fatalf("invalid level = %v", err)
	}
	if err := logger.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := logger.WriteDurable(context.Background(), InfoLevel, "stopped"); !errors.Is(err, ErrLoggerStopped) {
		t.Fatalf("stopped = %v", err)
	}
	logger, err = New()
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Stop()
	if err := logger.WriteDurable(context.Background(), InfoLevel, "no files"); !errors.Is(err, ErrNoDurableTarget) {
		t.Fatalf("no files = %v", err)
	}
}

type blockedDurableFormatter struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (f *blockedDurableFormatter) Format(entry *message.Entry) ([]byte, error) {
	f.once.Do(func() { close(f.started); <-f.release })
	return []byte(entry.Message + "\n"), nil
}

func TestWriteDurableCanceledAcceptedRequestStillDrains(t *testing.T) {
	requireDurablePlatform(t)
	worker, path := durableTestWorker(t, "cancel")
	fm := &blockedDurableFormatter{started: make(chan struct{}), release: make(chan struct{})}
	worker.SetFormatter(fm)
	logger, err := New(SetWorkerConfigs(worker))
	if err != nil {
		t.Fatal(err)
	}
	var release sync.Once
	unblock := func() { release.Do(func() { close(fm.release) }) }
	t.Cleanup(func() { unblock(); _ = logger.Stop() })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- logger.WriteDurable(ctx, InfoLevel, "accepted") }()
	<-fm.started
	cancel()
	err = <-done
	if !errors.Is(err, ErrDurabilityUnknown) || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled accepted write = %v", err)
	}
	stopDone := make(chan error, 1)
	go func() { stopDone <- logger.Stop() }()
	select {
	case <-stopDone:
		t.Fatal("Stop returned before accepted write finished")
	case <-time.After(20 * time.Millisecond):
	}
	unblock()
	select {
	case err := <-stopDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Stop blocked after canceled caller left")
	}
	if got := readDurableTestFile(t, path); got != "accepted\n" {
		t.Fatalf("accepted record lost: %q", got)
	}
}

func TestWriteDurableCanceledBeforeAdmissionWritesNothing(t *testing.T) {
	requireDurablePlatform(t)
	worker, path := durableTestWorker(t, "cancel")
	logger, err := New(SetWorkerConfigs(worker))
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := logger.WriteDurable(ctx, InfoLevel, "never accepted"); !errors.Is(err, context.Canceled) || errors.Is(err, ErrDurabilityUnknown) {
		t.Fatalf("canceled before admission = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("file was opened for canceled request: %v", err)
	}
}

func TestStopReleasesBlockedDurableEngineAdmission(t *testing.T) {
	requireDurablePlatform(t)
	worker, path := durableTestWorker(t, "engine-stop")
	fm := &blockedDurableFormatter{started: make(chan struct{}), release: make(chan struct{})}
	worker.SetFormatter(fm)
	logger, err := New(SetWorkerConfigs(worker))
	if err != nil {
		t.Fatal(err)
	}
	var release sync.Once
	unblock := func() { release.Do(func() { close(fm.release) }) }
	t.Cleanup(func() { unblock(); _ = logger.Stop() })
	first, second, third := make(chan error, 1), make(chan error, 1), make(chan error, 1)
	go func() { first <- logger.WriteDurable(context.Background(), InfoLevel, "first") }()
	<-fm.started
	go func() { second <- logger.WriteDurable(context.Background(), InfoLevel, "second") }()
	engine := logger.engine.(*ChannelEngine)
	deadline := time.Now().Add(5 * time.Second)
	for len(engine.workers[0].entryChan) != 1 {
		if time.Now().After(deadline) {
			t.Fatal("second request did not enter engine queue")
		}
		runtime.Gosched()
	}
	go func() { third <- logger.WriteDurable(context.Background(), InfoLevel, "third") }()
	stopped := make(chan error, 1)
	go func() { stopped <- logger.Stop() }()
	select {
	case err := <-third:
		if !errors.Is(err, ErrLoggerStopped) {
			t.Fatalf("blocked engine admission = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not release engine admission")
	}
	unblock()
	for _, result := range []chan error{first, second, stopped} {
		if err := <-result; err != nil {
			t.Fatal(err)
		}
	}
	if got := readDurableTestFile(t, path); got != "first\nsecond\n" {
		t.Fatalf("accepted engine records lost: %q", got)
	}
}

func TestPackageWriteDurableMetadataAndLegacyLogger(t *testing.T) {
	requireDurablePlatform(t)
	isolateGlobalLogger(t)
	worker, _ := durableTestWorker(t, "metadata")
	entries := make(chan message.Entry, 1)
	worker.SetFormatter(durableFormatFunc(func(entry *message.Entry) ([]byte, error) {
		entries <- *entry
		return []byte(entry.Message + "\n"), nil
	}))
	if err := InitLog(SetLoggerName("durable"), SetRecordCaller(0), SetTraceIDFunc(func(*message.Entry) string { return "trace" }), SetWorkerConfigs(worker)); err != nil {
		t.Fatal(err)
	}
	_, file, line, _ := runtime.Caller(0)
	err := WriteDurable(context.Background(), InfoLevel, "metadata")
	if err != nil {
		t.Fatal(err)
	}
	entry := <-entries
	if entry.LoggerName != "durable" || entry.TraceID != "trace" || entry.Time.IsZero() || entry.RoutineID == 0 || entry.Caller.File != file || entry.Caller.Line != line+1 {
		t.Fatalf("metadata = %+v, caller = %+v", entry, entry.Caller)
	}
	// Embedding the old interface deliberately hides the optional extension.
	type legacyLogger struct{ ILogger }
	current := GetLogger()
	SetLogger(legacyLogger{current})
	if err := WriteDurable(context.Background(), InfoLevel, "unsupported"); !errors.Is(err, ErrDurableUnsupported) {
		t.Fatalf("legacy logger = %v", err)
	}
}

func TestInitLogWaitsForDurableCall(t *testing.T) {
	requireDurablePlatform(t)
	isolateGlobalLogger(t)
	worker, path := durableTestWorker(t, "old")
	fm := &blockedDurableFormatter{started: make(chan struct{}), release: make(chan struct{})}
	var release sync.Once
	unblock := func() { release.Do(func() { close(fm.release) }) }
	t.Cleanup(unblock)
	worker.SetFormatter(fm)
	if err := InitLog(SetWorkerConfigs(worker)); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- WriteDurable(context.Background(), InfoLevel, "old record") }()
	<-fm.started
	// Replacing a logger using the same file must also preserve the in-flight
	// old call, with both file handlers coordinating the shared namespace.
	next := worker
	installed := make(chan error, 1)
	go func() { installed <- InitLog(SetWorkerConfigs(next)) }()
	select {
	case <-installed:
		t.Fatal("InitLog retired a logger with a live durable call")
	case <-time.After(20 * time.Millisecond):
	}
	unblock()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := <-installed; err != nil {
		t.Fatal(err)
	}
	if got := readDurableTestFile(t, path); got != "old record\n" {
		t.Fatalf("retired logger lost its accepted record: %q", got)
	}
}
