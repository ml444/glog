package handler

import (
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/internal/errorcallback"
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
	entryChan     chan message.Entry
	doneChan      chan struct{}
	workerDone    chan struct{}
	closeOnce     sync.Once
	closeErr      error
	sendMu        sync.RWMutex
	closed        uint32
	producersDone chan struct{}

	ErrorCallback  func(buf interface{}, err error)
	errorCallbacks errorcallback.Dispatcher
}

func NewFileHandler(cfg *FileHandlerConfig, fm formatter.IFormatter, ft filter.IFilter) (*FileHandler, error) {
	if cfg == nil {
		return nil, errors.New("file handler config is nil")
	}
	if cfg.BufferSize < 0 {
		return nil, errors.New("file handler buffer size must not be negative")
	}
	if fm == nil {
		return nil, errors.New("formatter is nil")
	}
	bulkWriteSize := cfg.BulkWriteSize
	if bulkWriteSize <= 0 {
		bulkWriteSize = 1024 * 1024
	}
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
		bulkWriteSize: bulkWriteSize,
		backpressure:  cfg.Backpressure.Normalize(BackpressureStrategyDrop),
		ErrorCallback: cfg.ErrCallback,
		entryChan:     make(chan message.Entry, cfg.BufferSize),
		doneChan:      make(chan struct{}),
		workerDone:    make(chan struct{}),
		producersDone: make(chan struct{}),
	}
	go h.flushWorker()
	return h, nil
}

func (h *FileHandler) flushWorker() {
	defer close(h.workerDone)
	var batch []byte
	for {
		select {
		case entry := <-h.entryChan:
			batch = h.flushBatch(batch[:0], entry)
		case <-h.doneChan:
			<-h.producersDone
			for {
				select {
				case entry := <-h.entryChan:
					batch = h.flushBatch(batch[:0], entry)
				default:
					return
				}
			}

		}
	}
}

func (h *FileHandler) flushBatch(batch []byte, first message.Entry) []byte {
	batch = h.appendEntry(batch, &first)
	batch = h.BulkFill(batch)
	if len(batch) == 0 {
		return batch
	}
	if err := h.realWrite(batch); err != nil {
		h.reportError(batch, err)
	}
	if cap(batch) > h.bulkWriteSize*2 {
		return nil
	}
	return batch
}

// BulkFill drains immediately available entries into buf until it reaches the
// configured batch size. It remains exported for compatibility with callers
// that previously used it to inspect batching behavior.
func (h *FileHandler) BulkFill(buf []byte) []byte {
	for len(buf) < h.bulkWriteSize {
		select {
		case entry := <-h.entryChan:
			buf = h.appendEntry(buf, &entry)
		default:
			return buf
		}
	}
	return buf
}

func (h *FileHandler) appendEntry(batch []byte, entry *message.Entry) []byte {
	var (
		formatted []byte
		err       error
	)
	switch fm := h.formatter.(type) {
	case *formatter.TextFormatter:
		formatted, err = fm.Append(batch, entry)
	case *formatter.JSONFormatter:
		formatted, err = fm.Append(batch, entry)
	case *formatter.XMLFormatter:
		formatted, err = fm.Append(batch, entry)
	default:
		formatted, err = h.formatter.Format(entry)
		if err == nil {
			formatted = append(batch, formatted...)
		}
	}
	if err != nil {
		h.reportError(entry, err)
		return batch
	}
	return formatted
}

func (h *FileHandler) reportError(value interface{}, err error) {
	if h.ErrorCallback != nil {
		if buf, ok := value.([]byte); ok {
			value = append([]byte(nil), buf...)
		}
		h.errorCallbacks.Report(h.ErrorCallback, value, err)
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
		file, needRotate, err = h.rotator.NeedRollover(buf)
		if err != nil {
			return err
		}
		if needRotate {
			return ErrFileSizeLimit
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

func (h *FileHandler) Emit(entry *message.Entry) error {
	h.sendMu.RLock()
	defer h.sendMu.RUnlock()
	if atomic.LoadUint32(&h.closed) != 0 {
		return ErrHandlerClosed
	}
	if entry == nil {
		return errors.New("log entry is nil")
	}
	if err := applyFilter(h.filter, entry); err != nil {
		return err
	}
	snapshot := *entry
	if entry.Caller != nil {
		caller := *entry.Caller
		snapshot.Caller = &caller
	}
	return h.enqueue(snapshot)
}

func (h *FileHandler) enqueue(entry message.Entry) error {
	switch h.backpressure.Strategy {
	case BackpressureStrategyBlock:
		select {
		case h.entryChan <- entry:
			h.stats.AddEnqueued()
			return nil
		case <-h.doneChan:
			return ErrHandlerClosed
		}
	case BackpressureStrategyTimeout:
		select {
		case h.entryChan <- entry:
			h.stats.AddEnqueued()
			return nil
		case <-h.doneChan:
			return ErrHandlerClosed
		default:
		}
		t := AcquireTimeoutTimer(h.backpressure.Timeout)
		defer ReleaseTimeoutTimer(t)
		select {
		case h.entryChan <- entry:
			h.stats.AddEnqueued()
			return nil
		case <-h.doneChan:
			return ErrHandlerClosed
		case <-t.C:
			h.stats.AddTimedOut()
			return ErrBackpressureTimeout
		}
	case BackpressureStrategySample:
		select {
		case h.entryChan <- entry:
			h.stats.AddEnqueued()
			return nil
		case <-h.doneChan:
			return ErrHandlerClosed
		default:
			if h.stats.ShouldSample(h.backpressure.SampleRate) {
				select {
				case h.entryChan <- entry:
					h.stats.AddEnqueued()
					h.stats.AddSampled()
					return nil
				case <-h.doneChan:
					return ErrHandlerClosed
				default:
				}
			}
			h.stats.AddDropped()
			return ErrBackpressureDropped
		}
	default:
		select {
		case h.entryChan <- entry:
			h.stats.AddEnqueued()
			return nil
		case <-h.doneChan:
			return ErrHandlerClosed
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
	h.closeOnce.Do(func() {
		atomic.StoreUint32(&h.closed, 1)
		close(h.doneChan)
		h.sendMu.Lock()
		close(h.producersDone)
		h.sendMu.Unlock()
		<-h.workerDone
		h.closeErr = h.rotator.Close()
	})
	return h.closeErr
}
