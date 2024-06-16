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
	Send(record *message.Record)
}

type Worker struct {
	handler        handler.IHandler
	recordChan     chan *message.Record
	onError        func(v interface{}, err error)
	levelThreshold Level
}

func (w *Worker) Run() {
	for record := range w.recordChan {

		if record.Level < w.levelThreshold {
			continue
		}
		err := w.handler.Emit(record)
		if err != nil {
			w.onError(record, err)
		}
	}
}

func (w *Worker) Close() {
	for len(w.recordChan) != 0 {
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
			println(fmt.Sprintf("err: %s, record: %+v \n", err.Error(), v))
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
			recordChan:     make(chan *message.Record, workerCfg.CacheSize),
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

func (e *ChannelEngine) Send(record *message.Record) {
	if e.stop {
		e.once.Do(e.closeAllChan)
		return
	}
	for _, worker := range e.workers {
		if record.Level < worker.levelThreshold {
			continue
		}
		worker.recordChan <- record
	}
}

func (e *ChannelEngine) closeAllChan() {
	for _, worker := range e.workers {
		close(worker.recordChan)
	}
}

func (e *ChannelEngine) Stop() (err error) {
	e.stop = true
	for _, worker := range e.workers {
		worker.Close()
	}
	return nil
}
