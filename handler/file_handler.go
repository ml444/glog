package handler

import (
	"errors"
	"io"
	"os"
	"time"

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
	bufChan       chan []byte
	doneChan      chan bool
	done          bool

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
		ErrorCallback: cfg.ErrCallback,
		bufChan:       make(chan []byte, 1024),
		doneChan:      make(chan bool, 100),
	}
	go h.flushWorker()
	return h, nil
}

func (h *FileHandler) flushWorker() {
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
					h.done = true
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
		if err == io.ErrShortWrite {
			for n < len(buf) {
				var x int
				x, err = file.Write(buf[n:])
				if err != nil {
					return err
				}
				n += x
			}
		}
		return err
	}
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
	if h.filter != nil {
		if ok := h.filter.Filter(entry); !ok {
			return filter.ErrFilterOut
		}
	}

	if h.formatter == nil {
		return errors.New("formatter is nil")
	}

	msgByte, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	// h.bufChan <- msgByte
	select {
	case h.bufChan <- msgByte:
	default:
		return errors.New("buffer is full")
	}
	return nil
}

func (h *FileHandler) Close() error {
	h.doneChan <- true
	for i := 0; i < 100; i++ {
		if h.done {
			break
		}
		<-time.After(1 * time.Millisecond)
	}
	return h.rotator.Close()
}
