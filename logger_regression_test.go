package log

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ml444/glog/message"
)

func isolateGlobalLogger(t *testing.T) {
	t.Helper()
	originalHolder := loggerState.Load().(*loggerHolder)
	originalConfig := Conf
	SetLogger(nil)
	Conf = NewDefaultConfig()
	t.Cleanup(func() {
		Stop()
		loggerMu.Lock()
		loggerState.Store(originalHolder)
		Conf = originalConfig
		loggerMu.Unlock()
	})
}

type closeAwareBuffer struct {
	mu         sync.Mutex
	buffer     bytes.Buffer
	closed     bool
	closeCount int
}

func (stream *closeAwareBuffer) Write(buf []byte) (int, error) {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if stream.closed {
		return 0, errors.New("stream closed")
	}
	return stream.buffer.Write(buf)
}

func (stream *closeAwareBuffer) Close() error {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	stream.closed = true
	stream.closeCount++
	return nil
}

func (stream *closeAwareBuffer) snapshot() (string, int) {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	return stream.buffer.String(), stream.closeCount
}

type closeAwareHandler struct {
	closeAwareBuffer
}

type nonComparableStream struct {
	*closeAwareBuffer
	labels []string
}

func (sink *closeAwareHandler) Emit(entry *message.Entry) error {
	_, err := sink.Write([]byte(entry.Message + "\n"))
	return err
}

func TestInitLogKeepsReusedResourcesOpen(t *testing.T) {
	for _, kind := range []string{"stream", "non-comparable stream", "custom handler"} {
		t.Run(kind, func(t *testing.T) {
			isolateGlobalLogger(t)
			sink := &closeAwareHandler{}
			worker := NewWorkerConfig(InfoLevel, 8)
			if kind == "stream" {
				worker.SetStreamHandlerConfig(&StreamHandlerConfig{Streamer: sink})
			} else if kind == "non-comparable stream" {
				worker.SetStreamHandlerConfig(&StreamHandlerConfig{Streamer: nonComparableStream{
					closeAwareBuffer: &sink.closeAwareBuffer,
					labels:           []string{"shared"},
				}})
			} else {
				worker.SetHandler(sink)
			}
			errorsSeen := make(chan error, 8)
			if err := InitLog(SetWorkerConfigs(worker), SetOnError(func(_ interface{}, err error) {
				errorsSeen <- err
			})); err != nil {
				t.Fatal(err)
			}
			if err := InitLog(SetLoggerName("renamed")); err != nil {
				t.Fatal(err)
			}
			if _, closeCount := sink.snapshot(); closeCount != 0 {
				t.Fatalf("shared resource closed during InitLog: %d", closeCount)
			}
			Info("after reinit")
			Stop()
			content, closeCount := sink.snapshot()
			if !strings.Contains(content, "after reinit") || closeCount != 1 {
				t.Fatalf("content = %q, close count = %d", content, closeCount)
			}
			select {
			case err := <-errorsSeen:
				t.Fatalf("unexpected output error: %v", err)
			default:
			}
		})
	}
}

func TestIndependentLoggersShareResourceLifetime(t *testing.T) {
	sink := &closeAwareHandler{}
	worker := NewWorkerConfig(InfoLevel, 8).SetHandler(sink)
	first, err := New(SetWorkerConfigs(worker))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first.Stop() })
	second, err := New(SetWorkerConfigs(worker))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Stop() })
	if err := first.Stop(); err != nil {
		t.Fatal(err)
	}
	if _, closeCount := sink.snapshot(); closeCount != 0 {
		t.Fatalf("resource closed while second logger was active: %d", closeCount)
	}
	second.Info("second logger")
	if err := second.Stop(); err != nil {
		t.Fatal(err)
	}
	if content, closeCount := sink.snapshot(); content != "second logger\n" || closeCount != 1 {
		t.Fatalf("content = %q, close count = %d", content, closeCount)
	}
}

func TestFailedInitLogReleasesOnlyNewResourceReferences(t *testing.T) {
	isolateGlobalLogger(t)
	sink := &closeAwareHandler{}
	worker := NewWorkerConfig(InfoLevel, 8).SetHandler(sink)
	if err := InitLog(SetWorkerConfigs(worker)); err != nil {
		t.Fatal(err)
	}
	original := GetLogger()
	invalidWorker := NewWorkerConfig(InfoLevel, 8).SetFileHandlerConfig(
		NewDefaultFileHandlerConfig(t.TempDir()).WithReMatch("["),
	)
	if err := InitLog(SetWorkerConfigs(worker, invalidWorker)); err == nil {
		t.Fatal("invalid regexp unexpectedly accepted")
	}
	if GetLogger() != original {
		t.Fatal("failed initialization replaced the active logger")
	}
	Info("still active")
	Stop()
	content, closeCount := sink.snapshot()
	if content != "still active\n" || closeCount != 1 {
		t.Fatalf("content = %q, close count = %d", content, closeCount)
	}
}

type delayedStringer struct {
	started chan struct{}
	release chan struct{}
}

func (value delayedStringer) String() string {
	close(value.started)
	<-value.release
	return "in-flight"
}

func TestInitLogWaitsForInflightCall(t *testing.T) {
	isolateGlobalLogger(t)
	oldSink, newSink := &closeAwareHandler{}, &closeAwareHandler{}
	if err := InitLog(SetWorkerConfigs(NewWorkerConfig(InfoLevel, 8).SetHandler(oldSink))); err != nil {
		t.Fatal(err)
	}
	oldHolder := loggerState.Load().(*loggerHolder)
	arg := delayedStringer{started: make(chan struct{}), release: make(chan struct{})}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(arg.release) }) }
	defer release()
	sendDone := make(chan struct{})
	go func() {
		Info(arg)
		close(sendDone)
	}()
	select {
	case <-arg.started:
	case <-time.After(time.Second):
		t.Fatal("log call did not start")
	}
	initDone := make(chan error, 1)
	go func() {
		initDone <- InitLog(SetWorkerConfigs(NewWorkerConfig(InfoLevel, 8).SetHandler(newSink)))
	}()
	deadline := time.After(time.Second)
	for loggerState.Load().(*loggerHolder) == oldHolder {
		select {
		case <-deadline:
			t.Fatal("new logger was not published")
		default:
			runtime.Gosched()
		}
	}
	select {
	case <-oldHolder.drained:
		t.Fatal("old logger drained while its call was still formatting")
	default:
	}
	Info("new logger")
	release()
	select {
	case <-sendDone:
	case <-time.After(time.Second):
		t.Fatal("in-flight log call did not finish")
	}
	select {
	case err := <-initDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("InitLog did not finish")
	}
	Stop()
	if content, closeCount := oldSink.snapshot(); content != "in-flight\n" || closeCount != 1 {
		t.Fatalf("old logger content = %q, close count = %d", content, closeCount)
	}
	if content, _ := newSink.snapshot(); content != "new logger\n" {
		t.Fatalf("new logger content = %q", content)
	}
}

func TestConcurrentInitLogPreservesAllMessages(t *testing.T) {
	isolateGlobalLogger(t)
	const producerCount, messagesPerProducer = 8, 200
	sink := &closeAwareHandler{}
	worker := NewWorkerConfig(InfoLevel, 16).SetHandler(sink)
	var outputErrors uint64
	if err := InitLog(SetWorkerConfigs(worker), SetOnError(func(interface{}, error) {
		atomic.AddUint64(&outputErrors, 1)
	})); err != nil {
		t.Fatal(err)
	}
	var producers sync.WaitGroup
	start := make(chan struct{})
	for producer := 0; producer < producerCount; producer++ {
		producers.Add(1)
		go func(producerID int) {
			defer producers.Done()
			<-start
			for sequence := 0; sequence < messagesPerProducer; sequence++ {
				Infof("producer=%d sequence=%d", producerID, sequence)
			}
		}(producer)
	}
	close(start)
	for iteration := 0; iteration < 16; iteration++ {
		if err := InitLog(SetLoggerName(fmt.Sprintf("logger-%d", iteration))); err != nil {
			t.Fatal(err)
		}
	}
	producers.Wait()
	Stop()
	content, closeCount := sink.snapshot()
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	seen := make(map[string]bool, len(lines))
	for _, line := range lines {
		if seen[line] {
			t.Fatalf("duplicate message: %q", line)
		}
		seen[line] = true
	}
	if len(seen) != producerCount*messagesPerProducer || closeCount != 1 || atomic.LoadUint64(&outputErrors) != 0 {
		t.Fatalf("messages = %d, close count = %d, output errors = %d", len(seen), closeCount, atomic.LoadUint64(&outputErrors))
	}
}

func TestInfofWithoutArgumentsPreservesFormatting(t *testing.T) {
	for _, template := range []string{"load: 100%%", "missing: %s", "literal", "trailing %"} {
		t.Run(template, func(t *testing.T) {
			sink := &closeAwareHandler{}
			logger, err := New(SetWorkerConfigs(NewWorkerConfig(InfoLevel, 8).SetHandler(sink)))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = logger.Stop() })
			logger.Infof(template)
			if err := logger.Stop(); err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf(template) + "\n"
			if content, _ := sink.snapshot(); content != want {
				t.Fatalf("got %q, want %q", content, want)
			}
		})
	}
}

func TestGlobalCallerLocationSurvivesLoggerLeases(t *testing.T) {
	isolateGlobalLogger(t)
	sink := &captureHandler{entries: make(chan *message.Entry, 1)}
	if err := InitLog(SetRecordCaller(0), SetWorkerConfigs(NewWorkerConfig(InfoLevel, 8).SetHandler(sink))); err != nil {
		t.Fatal(err)
	}
	_, expectedFile, expectedLine, _ := runtime.Caller(0)
	Info("caller")
	Stop()
	select {
	case entry := <-sink.entries:
		if entry.Caller == nil || entry.Caller.File != expectedFile || entry.Caller.Line != expectedLine+1 {
			t.Fatalf("caller = %+v, want %s:%d", entry.Caller, expectedFile, expectedLine+1)
		}
	default:
		t.Fatal("missing caller log")
	}
}
