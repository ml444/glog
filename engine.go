package log

import (
	"errors"
	"fmt"
	"sync"
	"time"

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
}

func (w *Worker) Run() {
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

func (w *Worker) Close() {
	for len(w.entryChan) != 0 {
		<-time.After(time.Millisecond * 1)
	}
	err := w.handler.Close()
	if err != nil {
		w.onError(nil, err)
	}
}

type ChannelEngine struct {
	workers []*Worker
	onError func(v interface{}, err error)

	once *sync.Once
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
		})
	}
	if len(workers) == 0 {
		return nil, errors.New("no Worker is configured")
	}

	return &ChannelEngine{
		workers: workers,
		onError: cfg.OnError,
		once:    &sync.Once{},
	}, nil
}

func (e *ChannelEngine) Start() error {
	for _, worker := range e.workers {
		go worker.Run()
	}
	return nil
}

func (e *ChannelEngine) Send(entry *message.Entry) {
	if e.stop {
		e.once.Do(e.closeAllChan)
		return
	}
	for _, worker := range e.workers {
		if entry.Level < worker.levelThreshold {
			continue
		}
		if e.stop {
			return
		}
		worker.entryChan <- entry
	}
}

func (e *ChannelEngine) closeAllChan() {
	for _, worker := range e.workers {
		close(worker.entryChan)
	}
}

func (e *ChannelEngine) Stop() (err error) {
	e.stop = true
	for _, worker := range e.workers {
		worker.Close()
	}
	return nil
}
