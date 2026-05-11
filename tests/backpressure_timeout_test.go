package tests

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	log "github.com/ml444/glog"
)

func TestWorkerBackpressureTimeoutStats(t *testing.T) {
	h := newBlockingHandler()
	var onErr uint64
	logger, err := log.NewLogger(&log.Config{
		LoggerLevel: log.DebugLevel,
		OnError: func(_ interface{}, err error) {
			if errors.Is(err, log.ErrBackpressureTimeout) {
				atomic.AddUint64(&onErr, 1)
			}
		},
		WorkerConfigList: []*log.WorkerConfig{
			log.NewWorkerConfig(log.PrintLevel, 1).
				SetBackpressure(log.BackpressureConfig{
					Strategy: log.BackpressureStrategyTimeout,
					Timeout:  15 * time.Millisecond,
				}).
				SetHandler(h),
		},
	})
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	logger.Info("first blocks handler")
	<-h.started

	logger.Info("fills queue")
	logger.Info("should time out enqueue")

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		st := logger.Stats()
		if len(st.Workers) == 1 && st.Workers[0].QueueBackpressure.TimedOut > 0 {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	st := logger.Stats()
	if len(st.Workers) != 1 {
		t.Fatalf("workers: %d", len(st.Workers))
	}
	if st.Workers[0].QueueBackpressure.TimedOut == 0 {
		t.Fatalf("expected TimedOut > 0, stats %+v", st.Workers[0].QueueBackpressure)
	}
	if atomic.LoadUint64(&onErr) == 0 {
		t.Fatal("expected OnError for timeout backpressure")
	}

	close(h.release)
	if err := logger.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}
