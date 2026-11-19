package handler

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
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
	entryChan     chan fileRequest
	path          string
	ownerKey      string
	owner         *fileCoordinator
	batchMu       sync.Mutex
	pending       *fileRequest
	pendingReady  chan struct{}
	writeFile     func(*os.File, []byte) (int, error)
	syncFile      func(*os.File) error
	syncDir       func(string) error
	writeErr      error // owned by the single file writer
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

type fileRequest struct {
	entry message.Entry
	done  chan error
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

	path := cfg.FileSuffix
	if cfg.FileName != "" {
		path = cfg.FileName + "." + path
	}
	path = filepath.Join(cfg.FileDir, path)
	ownerKey, owner := retainFileCoordinator(path)
	owner.mu.Lock()
	var rotator IRotator
	err := checkPendingSizeRotation(path, cfg.RotatorType)
	if err == nil {
		rotator, err = NewRotator(cfg)
	}
	owner.last = nil
	owner.mu.Unlock()
	if err != nil {
		releaseFileCoordinator(ownerKey, owner)
		return nil, err
	}
	h := &FileHandler{
		formatter:     fm,
		filter:        ft,
		rotator:       rotator,
		bulkWriteSize: bulkWriteSize,
		backpressure:  cfg.Backpressure.Normalize(BackpressureStrategyDrop),
		ErrorCallback: cfg.ErrCallback,
		entryChan:     make(chan fileRequest, cfg.BufferSize),
		path:          path,
		ownerKey:      ownerKey,
		owner:         owner,
		pendingReady:  make(chan struct{}, 1),
		writeFile:     func(file *os.File, buf []byte) (int, error) { return file.Write(buf) },
		syncFile:      syncDurableFile,
		syncDir:       syncDirectory,
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
		if request := h.takePending(); request != nil {
			batch = h.processRequest(batch[:0], *request)
			continue
		}
		select {
		case <-h.pendingReady:
			continue
		case request := <-h.entryChan:
			batch = h.processRequest(batch[:0], request)
		case <-h.doneChan:
			<-h.producersDone
			for {
				if request := h.takePending(); request != nil {
					batch = h.processRequest(batch[:0], *request)
					continue
				}
				select {
				case request := <-h.entryChan:
					batch = h.processRequest(batch[:0], request)
				default:
					return
				}
			}

		}
	}
}

func (h *FileHandler) takePending() *fileRequest {
	h.batchMu.Lock()
	request := h.pending
	h.pending = nil
	h.batchMu.Unlock()
	return request
}

func (h *FileHandler) processRequest(batch []byte, request fileRequest) []byte {
	if request.done == nil {
		return h.flushBatch(batch, request.entry)
	}
	formatted, err := h.formatEntry(nil, &request.entry)
	if err != nil {
		err = fileError(h.path, "format", err)
	} else if len(formatted) == 0 {
		err = fileError(h.path, "format", ErrEmptyLogRecord)
	} else {
		err = h.write(formatted, true)
	}
	request.done <- err
	return batch
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
	// This method is public. Register external consumers in the admission
	// barrier so Close cannot miss a reliable request moved into pending.
	h.sendMu.RLock()
	defer h.sendMu.RUnlock()
	if atomic.LoadUint32(&h.closed) != 0 {
		return buf
	}
	h.batchMu.Lock()
	defer h.batchMu.Unlock()
	if h.pending != nil {
		return buf
	}
	for len(buf) < h.bulkWriteSize {
		select {
		case request := <-h.entryChan:
			if request.done != nil {
				h.pending = &request
				select {
				case h.pendingReady <- struct{}{}:
				default:
				}
				return buf
			}
			buf = h.appendEntry(buf, &request.entry)
		default:
			return buf
		}
	}
	return buf
}

func (h *FileHandler) appendEntry(batch []byte, entry *message.Entry) []byte {
	formatted, err := h.formatEntry(batch, entry)
	if err != nil {
		h.reportError(entry, err)
		return batch
	}
	return formatted
}

func (h *FileHandler) formatEntry(batch []byte, entry *message.Entry) ([]byte, error) {
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
		return batch, err
	}
	return formatted, nil
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
	return h.write(buf, false)
}

func (h *FileHandler) write(buf []byte, durable bool) error {
	if h.owner != nil {
		h.owner.mu.Lock()
		defer h.owner.mu.Unlock()
		if h.owner.last != h {
			if err := refreshRotator(h.rotator); err != nil {
				return err
			}
			h.owner.last = h
		}
	}
	if h.writeErr != nil {
		return h.writeErr
	}
	var err error
	var needRotate bool
	var file *os.File
	file, needRotate, err = h.rotator.NeedRollover(buf)
	if err != nil {
		return fileError(h.path, "open", err)
	}
	if needRotate {
		file, err = h.rotator.DoRollover()
		if err != nil {
			return fileError(h.path, "rotate", err)
		}
		file, needRotate, err = h.rotator.NeedRollover(buf)
		if err != nil {
			return fileError(h.path, "rotate", err)
		}
		if needRotate {
			return fileError(h.path, "rotate", ErrFileSizeLimit)
		}
	}
	if file == nil {
		return fileError(h.path, "open", errors.New("file not open"))
	}
	write := h.writeFile
	if write == nil {
		write = func(file *os.File, buf []byte) (int, error) { return file.Write(buf) }
	}
	n, err := writeAll(file, buf, write)
	h.rotator.RecordBytesWritten(n)
	if err != nil {
		h.writeErr = fileError(file.Name(), "write", err)
		return h.writeErr
	}
	if durable {
		if err := h.syncFile(file); err != nil {
			h.writeErr = fileError(file.Name(), "sync", err)
			return h.writeErr
		}
		if err := h.syncDir(filepath.Dir(file.Name())); err != nil {
			h.writeErr = fileError(file.Name(), "sync-directory", err)
			return h.writeErr
		}
	}
	return nil
}

func writeAll(file *os.File, buf []byte, write func(*os.File, []byte) (int, error)) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := write(file, buf[total:])
		if n < 0 || n > len(buf)-total {
			return total, io.ErrShortWrite
		}
		total += n
		if err != nil {
			if !errors.Is(err, io.ErrShortWrite) || n == 0 || total == len(buf) {
				return total, err
			}
		}
		if n == 0 {
			return total, io.ErrNoProgress
		}
	}
	return total, nil
}

// Path is the configured file target; operation errors contain the actual path
// when a time-suffixed filename is in use.
func (h *FileHandler) Path() string { return h.path }

// EmitDurable bypasses ordinary drop/sample policies and waits for a separate
// write and strong synchronization. Once accepted, cancellation only stops the
// wait; the file writer still completes the request.
func (h *FileHandler) EmitDurable(ctx context.Context, entry *message.Entry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !durablePlatformSupported {
		return ErrDurableUnsupported
	}
	if atomic.LoadUint32(&h.closed) != 0 {
		return ErrHandlerClosed
	}
	if entry == nil {
		return errors.New("log entry is nil")
	}
	snapshot := *entry
	if entry.Caller != nil {
		caller := *entry.Caller
		snapshot.Caller = &caller
	}
	if err := applyFilter(h.filter, &snapshot); err != nil {
		return fileError(h.path, "filter", err)
	}
	done := make(chan error, 1)
	h.sendMu.RLock()
	if atomic.LoadUint32(&h.closed) != 0 {
		h.sendMu.RUnlock()
		return ErrHandlerClosed
	}
	if err := ctx.Err(); err != nil {
		h.sendMu.RUnlock()
		return err
	}
	select {
	case h.entryChan <- fileRequest{entry: snapshot, done: done}:
		h.stats.AddEnqueued()
		h.sendMu.RUnlock()
	case <-ctx.Done():
		h.sendMu.RUnlock()
		return ctx.Err()
	case <-h.doneChan:
		h.sendMu.RUnlock()
		return ErrHandlerClosed
	}
	select {
	case err := <-done:
		return err
	default:
	}
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return &DurabilityUnknownError{Cause: ctx.Err()}
	}
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
	request := fileRequest{entry: entry}
	switch h.backpressure.Strategy {
	case BackpressureStrategyBlock:
		select {
		case h.entryChan <- request:
			h.stats.AddEnqueued()
			return nil
		case <-h.doneChan:
			return ErrHandlerClosed
		}
	case BackpressureStrategyTimeout:
		select {
		case h.entryChan <- request:
			h.stats.AddEnqueued()
			return nil
		case <-h.doneChan:
			return ErrHandlerClosed
		default:
		}
		t := AcquireTimeoutTimer(h.backpressure.Timeout)
		defer ReleaseTimeoutTimer(t)
		select {
		case h.entryChan <- request:
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
		case h.entryChan <- request:
			h.stats.AddEnqueued()
			return nil
		case <-h.doneChan:
			return ErrHandlerClosed
		default:
			if h.stats.ShouldSample(h.backpressure.SampleRate) {
				select {
				case h.entryChan <- request:
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
		case h.entryChan <- request:
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
		h.owner.mu.Lock()
		h.closeErr = h.rotator.Close()
		if h.owner.last == h {
			h.owner.last = nil
		}
		h.owner.mu.Unlock()
		releaseFileCoordinator(h.ownerKey, h.owner)
		if h.closeErr == nil {
			h.closeErr = h.writeErr
		}
	})
	return h.closeErr
}
