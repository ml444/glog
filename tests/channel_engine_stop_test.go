package tests

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	log "github.com/ml444/glog"
	"github.com/ml444/glog/message"
)

type countingHandler struct {
	count uint64
}

func (h *countingHandler) Emit(_ *message.Entry) error {
	atomic.AddUint64(&h.count, 1)
	return nil
}

func (h *countingHandler) Close() error { return nil }

func TestChannelEngineStopDoesNotPanicDuringConcurrentSend(t *testing.T) {
	h := &countingHandler{}
	logger, err := log.NewLogger(&log.Config{
		LoggerLevel: log.DebugLevel,
		WorkerConfigList: []*log.WorkerConfig{
			log.NewWorkerConfig(log.PrintLevel, 16).
				SetBackpressureStrategy(log.BackpressureStrategyDrop).
				SetHandler(h),
		},
	})
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start
			for j := 0; j < 1000; j++ {
				logger.Infof("worker:%d msg:%d", id, j)
			}
		}(i)
	}
	close(start)

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := logger.Stop(); err != nil {
			t.Errorf("Stop: %v", err)
		}
		wg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop or concurrent senders did not finish")
	}
}

func TestChannelEngineStopReleasesBlockedBlockStrategySend(t *testing.T) {
	h := newBlockingHandler()
	logger, err := log.NewLogger(&log.Config{
		LoggerLevel: log.DebugLevel,
		WorkerConfigList: []*log.WorkerConfig{
			log.NewWorkerConfig(log.PrintLevel, 1).
				SetBackpressureStrategy(log.BackpressureStrategyBlock).
				SetHandler(h),
		},
	})
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	logger.Info("block worker")
	<-h.started

	logger.Info("fill queue")

	senderDone := make(chan struct{})
	go func() {
		defer close(senderDone)
		logger.Info("blocked sender")
	}()

	select {
	case <-senderDone:
		t.Fatal("sender finished before Stop released it")
	case <-time.After(25 * time.Millisecond):
	}

	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)
		if err := logger.Stop(); err != nil {
			t.Errorf("Stop: %v", err)
		}
	}()

	select {
	case <-senderDone:
	case <-time.After(time.Second):
		t.Fatal("blocked sender was not released after Stop")
	}

	close(h.release)

	select {
	case <-stopDone:
	case <-time.After(time.Second):
		t.Fatal("Stop did not finish after handler release")
	}
}
