package log

import (
	"errors"
	"fmt"
	"sync"

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
	// runDone is closed when Run returns (after entryChan is closed and drained).
	runDone chan struct{}
}

func (w *Worker) Run() {
	defer close(w.runDone)
	for entry := range w.entryChan {

		if entry.Level < w.levelThreshold {
			continue
		}
		err := w.handler.Emit(entry)
		if err != nil {
			w.onError(entry, err)
		}
	}
}

type ChannelEngine struct {
	workers []*Worker
	onError func(v interface{}, err error)

	mu   sync.RWMutex
	stop bool
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
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.stop {
		return
	}
	for _, worker := range e.workers {
		if entry.Level < worker.levelThreshold {
			continue
		}
		worker.entryChan <- entry
	}
}

func (e *ChannelEngine) Stop() (err error) {
	e.mu.Lock()
	if e.stop {
		e.mu.Unlock()
		return nil
	}
	e.stop = true
	for _, w := range e.workers {
		close(w.entryChan)
	}
	e.mu.Unlock()

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
