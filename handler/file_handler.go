package handler

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

type RotatorType int8

const (
	FileRotatorTypeTime        RotatorType = 1
	FileRotatorTypeSize        RotatorType = 2
	FileRotatorTypeTimeAndSize RotatorType = 3
)

type FileHandler struct {
	formatter formatter.IFormatter
	filter    filter.IFilter
	rotator   IRotator

	bulkWriteSize int
	backpressure  BackpressureConfig
	stats         BackpressureCounter
	bufChan    chan []byte
	doneChan   chan struct{}
	workerDone chan struct{}
	closeOnce  sync.Once

	ErrorCallback func(buf interface{}, err error)
}

func NewFileHandler(cfg *FileHandlerConfig, fm formatter.IFormatter, ft filter.IFilter) (*FileHandler, error) {
	// Redirects the standard error output to the specified file,
	// in order to preserve the panic information during panic.
	// rewriteStderr(handlerCfg.File.FileDir, config.GlobalConfig.LoggerName)

	rotator, err := NewRotator(cfg)
	if err != nil {
		return nil, err
	}
	h := &FileHandler{
		formatter:     fm,
		filter:        ft,
		rotator:       rotator,
		bulkWriteSize: cfg.BulkWriteSize,
		backpressure:  cfg.Backpressure.Normalize(BackpressureStrategyDrop),
		ErrorCallback: cfg.ErrCallback,
		bufChan:    make(chan []byte, cfg.BufferSize),
		doneChan:   make(chan struct{}),
		workerDone: make(chan struct{}),
	}
	go h.flushWorker()
	return h, nil
}

func (h *FileHandler) flushWorker() {
	defer close(h.workerDone)
	for {
		select {
		case b := <-h.bufChan:
			buf := h.BulkFill(b)
			err := h.realWrite(buf)
			if err != nil && h.ErrorCallback != nil {
				h.ErrorCallback(buf, err)
			}
		case <-h.doneChan:
			for {
				select {
				case b := <-h.bufChan: // Flush channel data into storage
					buf := h.BulkFill(b)
					err := h.realWrite(buf)
					if err != nil && h.ErrorCallback != nil {
						h.ErrorCallback(buf, err)
					}
				default:
					return
				}
			}

		}
	}
}

func (h *FileHandler) realWrite(buf []byte) error {
	var err error
	var needRotate bool
	var file *os.File
	file, needRotate, err = h.rotator.NeedRollover(buf)
	if err != nil {
		return err
	}
	if needRotate {
		file, err = h.rotator.DoRollover()
		if err != nil {
			return err
		}
	}
	if file == nil {
		return errors.New("file not open")
	}
	n, err := file.Write(buf)
	if err != nil {
		if !errors.Is(err, io.ErrShortWrite) {
			return err
		}
		for n < len(buf) {
			var x int
			x, err = file.Write(buf[n:])
			if err != nil {
				return err
			}
			n += x
		}
	}
	h.rotator.RecordBytesWritten(n)
	return nil
}

func (h *FileHandler) BulkFill(buf []byte) []byte {
	total := len(buf)
	for {
		select {
		case more := <-h.bufChan:
			buf = append(buf, more...)
			total += len(more)
			if total >= h.bulkWriteSize {
				return buf
			}
		default:
			return buf
		}
	}
}

func (h *FileHandler) Emit(entry *message.Entry) error {
	if err := applyFilter(h.filter, entry); err != nil {
		return err
	}

	if h.formatter == nil {
		return errors.New("formatter is nil")
	}

	msgByte, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	return h.enqueue(msgByte)
}

func (h *FileHandler) enqueue(msg []byte) error {
	switch h.backpressure.Strategy {
	case BackpressureStrategyBlock:
		h.bufChan <- msg
		h.stats.AddEnqueued()
		return nil
	case BackpressureStrategyTimeout:
		t := AcquireTimeoutTimer(h.backpressure.Timeout)
		defer ReleaseTimeoutTimer(t)
		select {
		case h.bufChan <- msg:
			h.stats.AddEnqueued()
			return nil
		case <-t.C:
			h.stats.AddTimedOut()
			return ErrBackpressureTimeout
		}
	case BackpressureStrategySample:
		select {
		case h.bufChan <- msg:
			h.stats.AddEnqueued()
			return nil
		default:
			if h.stats.AllowSample(h.backpressure.SampleRate) {
				h.bufChan <- msg
				h.stats.AddEnqueued()
				return nil
			}
			h.stats.AddDropped()
			return ErrBackpressureDropped
		}
	default:
		select {
		case h.bufChan <- msg:
			h.stats.AddEnqueued()
			return nil
		default:
			h.stats.AddDropped()
			return ErrBackpressureDropped
		}
	}
}

func (h *FileHandler) BackpressureStats() BackpressureStats {
	return h.stats.Snapshot()
}

func (h *FileHandler) Close() error {
	var err error
	h.closeOnce.Do(func() {
		close(h.doneChan)
		<-h.workerDone
		err = h.rotator.Close()
	})
	return err
}
