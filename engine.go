package log

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/internal/errorcallback"
	"github.com/ml444/glog/message"
)

type IEngine interface {
	Start() error
	Stop() error
	Send(entry *message.Entry)
}

type Worker struct {
	handler        handler.IHandler
	fileHandler    *handler.FileHandler
	entryChan      chan engineRequest
	onError        func(v interface{}, err error)
	errorCallbacks errorcallback.Dispatcher
	levelThreshold Level
	backpressure   BackpressureConfig
	stats          BackpressureCounter
	stopChan       chan struct{}
	// runDone is closed when Run returns after stopChan is closed and entryChan is drained.
	runDone       chan struct{}
	producersDone <-chan struct{}
}

type engineRequest struct {
	entry *message.Entry
	done  chan error
}

func (w *Worker) Run() {
	defer close(w.runDone)
	for {
		select {
		case request := <-w.entryChan:
			w.process(request)
		case <-w.stopChan:
			<-w.producersDone
			w.drain()
			return
		}
	}
}

func (w *Worker) drain() {
	for {
		select {
		case request := <-w.entryChan:
			w.process(request)
		default:
			return
		}
	}
}

func (w *Worker) process(request engineRequest) {
	if request.done != nil {
		// Admission into the engine commits us to finishing this request, even
		// when its caller has stopped waiting. Stop drains these requests too.
		request.done <- w.fileHandler.EmitDurable(context.Background(), request.entry)
		return
	}
	w.emit(request.entry)
}

func (w *Worker) emit(entry *message.Entry) {
	if entry.Level < w.levelThreshold {
		return
	}
	err := w.handler.Emit(entry)
	if err != nil && !errors.Is(err, filter.ErrFilterOut) {
		w.errorCallbacks.Report(w.onError, entry, err)
	}
}

type ChannelEngine struct {
	workers       []*Worker
	onError       func(v interface{}, err error)
	stop          uint32
	sendMu        sync.RWMutex
	stopOnce      sync.Once
	stopErr       error
	producersDone chan struct{}
}

type workerFailure struct {
	worker *Worker
	err    error
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
	producersDone := make(chan struct{})
	for _, workerCfg := range cfg.WorkerConfigList {
		h, err := newHandler(workerCfg)
		if err != nil {
			for _, worker := range workers {
				_ = worker.handler.Close()
			}
			return nil, err
		}
		var fileHandler *handler.FileHandler
		if workerCfg.CustomHandler == nil {
			fileHandler, _ = h.(*handler.FileHandler)
		}
		workers = append(workers, &Worker{
			handler:        h,
			fileHandler:    fileHandler,
			entryChan:      make(chan engineRequest, workerCfg.CacheSize),
			onError:        cfg.OnError,
			levelThreshold: workerCfg.Level,
			backpressure:   workerCfg.Backpressure,
			stopChan:       make(chan struct{}),
			runDone:        make(chan struct{}),
			producersDone:  producersDone,
		})
	}
	if len(workers) == 0 {
		return nil, errors.New("no Worker is configured")
	}

	return &ChannelEngine{
		workers:       workers,
		onError:       cfg.OnError,
		producersDone: producersDone,
	}, nil
}

func (e *ChannelEngine) Start() error {
	for _, worker := range e.workers {
		go worker.Run()
	}
	return nil
}

func (e *ChannelEngine) Send(entry *message.Entry) {
	var failures []workerFailure
	e.sendMu.RLock()
	if atomic.LoadUint32(&e.stop) == 1 {
		e.sendMu.RUnlock()
		return
	}
	for _, worker := range e.workers {
		if entry.Level < worker.levelThreshold {
			continue
		}
		if err := worker.send(entry); err != nil {
			failures = append(failures, workerFailure{worker: worker, err: err})
		}
	}
	e.sendMu.RUnlock()
	for _, failure := range failures {
		failure.worker.onError(entry, failure.err)
	}
}

func (w *Worker) Send(entry *message.Entry) {
	if err := w.send(entry); err != nil {
		w.onError(entry, err)
	}
}

func (w *Worker) send(entry *message.Entry) error {
	request := engineRequest{entry: entry}
	switch w.backpressure.Strategy {
	case BackpressureStrategyDrop:
		select {
		case <-w.stopChan:
			w.stats.AddDropped()
			return nil
		default:
		}
		select {
		case w.entryChan <- request:
			w.stats.AddEnqueued()
		case <-w.stopChan:
			w.stats.AddDropped()
		default:
			w.stats.AddDropped()
			return handler.ErrBackpressureDropped
		}
	case BackpressureStrategyTimeout:
		select {
		case w.entryChan <- request:
			w.stats.AddEnqueued()
			return nil
		case <-w.stopChan:
			w.stats.AddDropped()
			return nil
		default:
		}
		t := handler.AcquireTimeoutTimer(w.backpressure.Timeout)
		defer handler.ReleaseTimeoutTimer(t)
		select {
		case w.entryChan <- request:
			w.stats.AddEnqueued()
		case <-w.stopChan:
			w.stats.AddDropped()
		case <-t.C:
			w.stats.AddTimedOut()
			return handler.ErrBackpressureTimeout
		}
	case BackpressureStrategySample:
		select {
		case <-w.stopChan:
			w.stats.AddDropped()
			return nil
		default:
		}
		select {
		case w.entryChan <- request:
			w.stats.AddEnqueued()
		case <-w.stopChan:
			w.stats.AddDropped()
		default:
			if w.stats.ShouldSample(w.backpressure.SampleRate) {
				select {
				case w.entryChan <- request:
					w.stats.AddEnqueued()
					w.stats.AddSampled()
				case <-w.stopChan:
					w.stats.AddDropped()
				default:
					w.stats.AddDropped()
					return handler.ErrBackpressureDropped
				}
				return nil
			}
			w.stats.AddDropped()
			return handler.ErrBackpressureDropped
		}
	default:
		select {
		case w.entryChan <- request:
			w.stats.AddEnqueued()
		case <-w.stopChan:
			w.stats.AddDropped()
		}
	}
	return nil
}

func (e *ChannelEngine) Stop() error {
	var closeErrors []workerFailure
	e.stopOnce.Do(func() {
		atomic.StoreUint32(&e.stop, 1)
		for _, worker := range e.workers {
			close(worker.stopChan)
		}
		e.sendMu.Lock()
		close(e.producersDone)
		e.sendMu.Unlock()
		for _, worker := range e.workers {
			<-worker.runDone
		}
		for _, worker := range e.workers {
			if err := worker.handler.Close(); err != nil {
				if e.stopErr == nil {
					e.stopErr = err
				}
				closeErrors = append(closeErrors, workerFailure{worker: worker, err: err})
			}
		}
	})
	for _, failure := range closeErrors {
		failure.worker.onError(nil, failure.err)
	}
	return e.stopErr
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
