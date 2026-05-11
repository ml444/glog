package tests

import (
	"sync/atomic"
	"testing"

	log "github.com/ml444/glog"
	"github.com/ml444/glog/message"
)

type rejectAllFilter struct{}

func (rejectAllFilter) Filter(*message.Entry) bool { return false }

func TestFilterOutDoesNotInvokeOnError(t *testing.T) {
	var errCount uint64
	logger, err := log.NewLogger(&log.Config{
		LoggerLevel: log.DebugLevel,
		OnError: func(_ interface{}, _ error) {
			atomic.AddUint64(&errCount, 1)
		},
		WorkerConfigList: []*log.WorkerConfig{
			log.NewWorkerConfig(log.PrintLevel, 64).
				SetTextFormatterConfig(log.NewDefaultTextFormatterConfig()).
				SetFilter(rejectAllFilter{}),
		},
	})
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for i := 0; i < 50; i++ {
		logger.Infof("n:%d", i)
	}
	if err := logger.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if atomic.LoadUint64(&errCount) != 0 {
		t.Fatalf("OnError called %d times, want 0 for filter-out", errCount)
	}
}
