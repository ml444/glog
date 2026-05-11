package log

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/message"
)

type IEngine interface {
	Start() error
	Stop() error
	Send(entry *message.Entry)
}

type Worker struct {
	handler        handler.IHandler
	entryChan      chan *message.Entry
	onError        func(v interface{}, err error)
	levelThreshold Level
	backpressure   BackpressureConfig
	stats          BackpressureCounter
	stopChan       chan struct{}
	// runDone is closed when Run returns after stopChan is closed and entryChan is drained.
	runDone chan struct{}
}

func (w *Worker) Run() {
	defer close(w.runDone)
	for {
		select {
		case entry := <-w.entryChan:
			w.emit(entry)
		case <-w.stopChan:
			w.drain()
			return
		}
	}
}

func (w *Worker) drain() {
	for {
		select {
		case entry := <-w.entryChan:
			w.emit(entry)
		default:
			return
		}
	}
}

func (w *Worker) emit(entry *message.Entry) {
	if entry.Level < w.levelThreshold {
		return
	}
	err := w.handler.Emit(entry)
	if err != nil && !errors.Is(err, filter.ErrFilterOut) {
		w.onError(entry, err)
	}
}

type ChannelEngine struct {
	workers []*Worker
	onError func(v interface{}, err error)
	stop    uint32
}

type WorkerStats struct {
	Level               Level
	QueueBackpressure   BackpressureStats
	HandlerBackpressure BackpressureStats
}

type LoggerStats struct {
	Workers []WorkerStats
}

func NewChannelEngine(cfg *Config) (*ChannelEngine, error) {
	if cfg.OnError == nil {
		cfg.OnError = func(v interface{}, err error) {
			println(fmt.Sprintf("err: %s, entry: %+v \n", err.Error(), v))
		}
	}
	var workers []*Worker
	for _, workerCfg := range cfg.WorkerConfigList {
		h, err := newHandler(workerCfg)
		if err != nil {
			return nil, err
		}
		workers = append(workers, &Worker{
			handler:        h,
			entryChan:      make(chan *message.Entry, workerCfg.CacheSize),
			onError:        cfg.OnError,
			levelThreshold: workerCfg.Level,
			backpressure:   workerCfg.Backpressure,
			stopChan:       make(chan struct{}),
			runDone:        make(chan struct{}),
		})
	}
	if len(workers) == 0 {
		return nil, errors.New("no Worker is configured")
	}

	return &ChannelEngine{
		workers: workers,
		onError: cfg.OnError,
	}, nil
}

func (e *ChannelEngine) Start() error {
	for _, worker := range e.workers {
		go worker.Run()
	}
	return nil
}

func (e *ChannelEngine) Send(entry *message.Entry) {
	if atomic.LoadUint32(&e.stop) == 1 {
		return
	}
	for _, worker := range e.workers {
		if entry.Level < worker.levelThreshold {
			continue
		}
		worker.Send(entry)
	}
}

func (w *Worker) Send(entry *message.Entry) {
	switch w.backpressure.Strategy {
	case BackpressureStrategyDrop:
		select {
		case <-w.stopChan:
			w.stats.AddDropped()
			return
		default:
		}
		select {
		case w.entryChan <- entry:
			w.stats.AddEnqueued()
		case <-w.stopChan:
			w.stats.AddDropped()
		default:
			w.stats.AddDropped()
			w.onError(entry, handler.ErrBackpressureDropped)
		}
	case BackpressureStrategyTimeout:
		t := handler.AcquireTimeoutTimer(w.backpressure.Timeout)
		defer handler.ReleaseTimeoutTimer(t)
		select {
		case w.entryChan <- entry:
			w.stats.AddEnqueued()
		case <-w.stopChan:
			w.stats.AddDropped()
		case <-t.C:
			w.stats.AddTimedOut()
			w.onError(entry, handler.ErrBackpressureTimeout)
		}
	case BackpressureStrategySample:
		select {
		case <-w.stopChan:
			w.stats.AddDropped()
			return
		default:
		}
		select {
		case w.entryChan <- entry:
			w.stats.AddEnqueued()
		case <-w.stopChan:
			w.stats.AddDropped()
		default:
			if w.stats.AllowSample(w.backpressure.SampleRate) {
				select {
				case w.entryChan <- entry:
					w.stats.AddEnqueued()
				case <-w.stopChan:
					w.stats.AddDropped()
				}
				return
			}
			w.stats.AddDropped()
			w.onError(entry, handler.ErrBackpressureDropped)
		}
	default:
		select {
		case w.entryChan <- entry:
			w.stats.AddEnqueued()
		case <-w.stopChan:
			w.stats.AddDropped()
		}
	}
}

func (e *ChannelEngine) Stop() (err error) {
	if !atomic.CompareAndSwapUint32(&e.stop, 0, 1) {
		return nil
	}
	for _, w := range e.workers {
		close(w.stopChan)
	}

	for _, w := range e.workers {
		<-w.runDone
	}
	for _, w := range e.workers {
		if cerr := w.handler.Close(); cerr != nil {
			w.onError(nil, cerr)
		}
	}
	return nil
}

func (e *ChannelEngine) Stats() LoggerStats {
	stats := LoggerStats{Workers: make([]WorkerStats, 0, len(e.workers))}
	for _, w := range e.workers {
		workerStats := WorkerStats{
			Level:             w.levelThreshold,
			QueueBackpressure: w.stats.Snapshot(),
		}
		if provider, ok := w.handler.(handler.BackpressureStatsProvider); ok {
			workerStats.HandlerBackpressure = provider.BackpressureStats()
		}
		stats.Workers = append(stats.Workers, workerStats)
	}
	return stats
}
