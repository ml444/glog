package log

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ml444/glog/message"
)

type shutdownGateHandler struct {
	once       sync.Once
	started    chan struct{}
	release    chan struct{}
	closeCount uint32
	closeErr   error
}

func (sink *shutdownGateHandler) Emit(*message.Entry) error {
	sink.once.Do(func() { close(sink.started) })
	<-sink.release
	return nil
}

func TestBackpressureCallbackCanStopLogger(t *testing.T) {
	sink := &shutdownGateHandler{started: make(chan struct{}), release: make(chan struct{})}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(sink.release) }) }
	var logger *Logger
	var err error
	callbackResult := make(chan error, 1)
	logger, err = New(SetWorkerConfigs(NewWorkerConfig(InfoLevel, 1).
		SetBackpressureStrategy(BackpressureStrategyDrop).SetHandler(sink)), SetOnError(func(interface{}, error) {
		release()
		callbackResult <- logger.Stop()
	}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		release()
		_ = logger.Stop()
	})
	logger.Info("first")
	select {
	case <-sink.started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	logger.Info("queued")
	sent := make(chan struct{})
	go func() {
		logger.Info("dropped")
		close(sent)
	}()
	select {
	case <-sent:
	case <-time.After(time.Second):
		t.Fatal("backpressure callback deadlocked while stopping")
	}
	if err := <-callbackResult; err != nil {
		t.Fatal(err)
	}
}

func (sink *shutdownGateHandler) Close() error {
	atomic.AddUint32(&sink.closeCount, 1)
	return sink.closeErr
}

func TestConcurrentStopsWaitAndReturnCloseError(t *testing.T) {
	want := errors.New("sync failed")
	sink := &shutdownGateHandler{started: make(chan struct{}), release: make(chan struct{}), closeErr: want}
	var callbacks uint32
	logger, err := New(SetWorkerConfigs(NewWorkerConfig(InfoLevel, 1).SetHandler(sink)), SetOnError(func(interface{}, error) {
		atomic.AddUint32(&callbacks, 1)
	}))
	if err != nil {
		t.Fatal(err)
	}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(sink.release) }) }
	t.Cleanup(func() {
		release()
		_ = logger.Stop()
	})
	logger.Info("in progress")
	select {
	case <-sink.started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	results := make(chan error, 2)
	go func() { results <- logger.Stop() }()
	go func() { results <- logger.Stop() }()
	select {
	case err := <-results:
		t.Fatalf("Stop returned before the worker finished: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	release()
	for count := 0; count < 2; count++ {
		select {
		case err := <-results:
			if !errors.Is(err, want) {
				t.Fatalf("Stop error = %v, want %v", err, want)
			}
		case <-time.After(time.Second):
			t.Fatal("Stop did not finish")
		}
	}
	if err := logger.Stop(); !errors.Is(err, want) {
		t.Fatalf("repeated Stop error = %v, want %v", err, want)
	}
	if got := atomic.LoadUint32(&sink.closeCount); got != 1 {
		t.Fatalf("Close calls = %d, want 1", got)
	}
	if got := atomic.LoadUint32(&callbacks); got != 1 {
		t.Fatalf("OnError calls = %d, want 1", got)
	}
}

func TestShutdownErrorCallbackCanObserveStopResult(t *testing.T) {
	want := errors.New("close failed")
	sink := &shutdownGateHandler{closeErr: want}
	var logger *Logger
	observed := make(chan error, 1)
	var err error
	logger, err = New(SetWorkerConfigs(NewWorkerConfig(InfoLevel, 1).SetHandler(sink)), SetOnError(func(interface{}, error) {
		observed <- logger.Stop()
	}))
	if err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() { finished <- logger.Stop() }()
	select {
	case err := <-finished:
		if !errors.Is(err, want) {
			t.Fatalf("Stop error = %v, want %v", err, want)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown callback deadlocked")
	}
	if err := <-observed; !errors.Is(err, want) {
		t.Fatalf("callback observed %v, want %v", err, want)
	}
}

func TestDynamicLevelChangesAreConcurrentSafe(t *testing.T) {
	sink := &closeAwareHandler{}
	logger, err := New(SetLoggerLevel(InfoLevel), SetWorkerConfigs(NewDefaultStdoutWorkerConfig().SetHandler(sink)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = logger.Stop() })
	var workers sync.WaitGroup
	start := make(chan struct{})
	workers.Add(3)
	go func() {
		defer workers.Done()
		<-start
		for iteration := 0; iteration < 1000; iteration++ {
			logger.SetLevel(DebugLevel)
			logger.SetLevel(InfoLevel)
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for iteration := 0; iteration < 1000; iteration++ {
			if !logger.Enabled(InfoLevel) {
				t.Error("Info became disabled")
				return
			}
			_ = logger.GetLevel()
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for iteration := 0; iteration < 1000; iteration++ {
			logger.Info("info")
		}
	}()
	close(start)
	workers.Wait()
	logger.SetLevel(DebugLevel)
	logger.Debug("debug after change")
	if err := logger.Stop(); err != nil {
		t.Fatal(err)
	}
	content, _ := sink.snapshot()
	if strings.Count(content, "info\n") != 1000 || !strings.Contains(content, "debug after change\n") {
		t.Fatal("level changes filtered valid messages")
	}
}
