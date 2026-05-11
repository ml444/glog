package log

import (
	"testing"
	"time"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/message"
)

func BenchmarkLoggerInfoDiscard(b *testing.B) {
	cfg := &Config{
		LoggerLevel: DebugLevel,
		WorkerConfigList: []*WorkerConfig{
			NewWorkerConfig(PrintLevel, 4096).
				SetBackpressure(handler.BackpressureConfig{
					Strategy: handler.BackpressureStrategyDrop,
				}).
				SetHandler(&noopHandler{}),
		},
	}
	logger, err := NewLogger(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = logger.Stop() }()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

type noopHandler struct{}

func (noopHandler) Emit(*message.Entry) error { return nil }
func (noopHandler) Close() error              { return nil }

func BenchmarkTextFormatterFormat(b *testing.B) {
	fm := formatter.NewTextFormatter(*NewDefaultTextFormatterConfig())
	entry := &message.Entry{
		Message: "hello world",
		Level:   InfoLevel,
		Time:    time.Now(),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fm.Format(entry)
		if err != nil {
			b.Fatal(err)
		}
	}
}
