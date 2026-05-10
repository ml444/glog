package tests

import (
	"sync"
	"sync/atomic"
	"testing"

	log "github.com/ml444/glog"
	"github.com/ml444/glog/message"
)

type blockingHandler struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
	count   uint64
}

func newBlockingHandler() *blockingHandler {
	return &blockingHandler{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (h *blockingHandler) Emit(_ *message.Entry) error {
	h.once.Do(func() {
		close(h.started)
	})
	<-h.release
	atomic.AddUint64(&h.count, 1)
	return nil
}

func (h *blockingHandler) Close() error { return nil }

func TestWorkerBackpressureDropStats(t *testing.T) {
	h := newBlockingHandler()
	var errCount uint64
	logger, err := log.NewLogger(&log.Config{
		LoggerLevel: log.DebugLevel,
		OnError: func(_ interface{}, _ error) {
			atomic.AddUint64(&errCount, 1)
		},
		WorkerConfigList: []*log.WorkerConfig{
			log.NewWorkerConfig(log.PrintLevel, 1).
				SetBackpressureStrategy(log.BackpressureStrategyDrop).
				SetHandler(h),
		},
	})
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	logger.Info("block worker")
	<-h.started

	for i := 0; i < 100; i++ {
		logger.Infof("msg:%d", i)
	}

	stats := logger.Stats()
	if len(stats.Workers) != 1 {
		t.Fatalf("worker stats len = %d, want 1", len(stats.Workers))
	}
	queue := stats.Workers[0].QueueBackpressure
	if queue.Dropped == 0 {
		t.Fatalf("dropped = 0, want drops under pressure; stats: %+v", queue)
	}
	if queue.Enqueued == 0 {
		t.Fatalf("enqueued = 0, want at least one queued message; stats: %+v", queue)
	}
	if got := atomic.LoadUint64(&errCount); got == 0 {
		t.Fatal("OnError was not called for dropped messages")
	}

	close(h.release)
	if err := logger.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}
