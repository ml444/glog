package handlers

import (
	"errors"
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/filters"
	"github.com/ml444/glog/formatters"
	"github.com/ml444/glog/message"
	"io"
	"os"
	"sync"
	"time"
)


type FileHandler struct {
	formatter formatters.IFormatter
	filter    filters.IFilter
	rotator   IRotator

	bufChan      chan []byte
	flushChan    chan bool
	workerDone   bool
	workerDoneMu sync.Mutex

	ErrorCallback func(err error)
}

func NewFileHandler(handlerCfg config.FileHandlerConfig, formatter formatters.IFormatter, filter filters.IFilter) (*FileHandler, error) {
	rotator, err := GetRotator4Config(&handlerCfg)
	if err != nil {
		return nil, err
	}
	h := &FileHandler{
		formatter:   formatter,
		filter:      filter,
		rotator:     rotator,
	}
	h.init()
	return h, nil
}

func (h *FileHandler) init() {
	h.bufChan = make(chan []byte, 1024)
	h.flushChan = make(chan bool, 100)
	go h.flushWorker()
	return
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
				x, err := file.Write(buf[n:])
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


func (h *FileHandler) flushWorker() {
	for {
		select {
		case buf := <-h.bufChan:
			var bb []byte
			var total int
			for {
				select {
				case more := <-h.bufChan:
					if len(bb) == 0 {
						bb = append(bb, buf...)
						total += len(buf)
					}
					bb = append(bb, more...)
					total += len(more)
					if total >= 1024*1024 {
						goto OUT
					}
				default:
					goto OUT
				}
			}
		OUT:
			if len(bb) == 0 {
				err := h.realWrite(buf)
				if err != nil && h.ErrorCallback != nil {
					h.ErrorCallback(err)
				}
			} else {
				err := h.realWrite(bb)
				if err != nil && h.ErrorCallback != nil {
					h.ErrorCallback(err)
				}
			}
		case <-h.flushChan:
			ok := true
			for ok {
				select {
				case buf := <-h.bufChan:
					_ = h.realWrite(buf)
				default:
					ok = false
					break
				}
			}
			h.workerDoneMu.Lock()
			h.workerDone = true
			h.workerDoneMu.Unlock()
			return
		}
	}
}

func (h *FileHandler) getWorkerDone() bool {
	h.workerDoneMu.Lock()
	res := h.workerDone
	h.workerDoneMu.Unlock()
	return res
}

func (h *FileHandler) Flush() {
	select {
	case h.flushChan <- true: // send
	default: // channel full
	}
	for i := 0; i < 100; i++ {
		if h.getWorkerDone() {
			break
		}
		time.Sleep(50 * time.Microsecond)
	}
}

func (h *FileHandler) Emit(entry *message.Entry) error {
	if h.filter != nil {
		if ok := h.filter.Filter(entry); !ok {
			return nil
			//return errors.New(fmt.Sprintf("Filter out this msg: %v", entry))
		}
	}

	if h.formatter != nil {
		msgByte, err := h.formatter.Format(entry)
		if err != nil {
			return err
		}

		h.bufChan <- msgByte
		//select {
		//case h.bufChan <- msgByte:
		//default:
		//	println("waring: buffer channel isFull")
		//	return errors.New("buffer channel isFull")
		//}
	}
	return nil
}

func (h *FileHandler) Sync() error {
	return h.rotator.Close()
}



