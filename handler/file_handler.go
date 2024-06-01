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

const (
	DefaultMaxFileSize   = 10 * 1024 * 1024 // 默认文件大小
	DefaultBackupCount   = 10               // 默认备份数量
	DefaultBulkWriteSize = 10 * 1024 * 1024 // 默认写入缓存大小
	DefaultInterval      = 60 * 60          // 默认轮转间隔
	DefaultFileSuffix    = "log"            // 默认文件后缀
)

type RotatorType int

const (
	FileRotatorTypeTime        RotatorType = 1
	FileRotatorTypeSize        RotatorType = 2
	FileRotatorTypeTimeAndSize RotatorType = 3
)

const (
	FileRotatorSuffixFmt1 = "20060102150405"
	FileRotatorSuffixFmt2 = "2006-01-02T15-04-05"
	FileRotatorSuffixFmt3 = "2006-01-02_15-04-05"
)

const (
	FileRotatorReMatch1 = "^\\d{14}(\\.\\w+)?$"
	FileRotatorReMatch2 = "^\\d{4}-\\d{2}-\\d{2}T\\d{2}-\\d{2}-\\d{2}(\\.\\w+)?$"
	FileRotatorReMatch3 = "^\\d{4}-\\d{2}-\\d{2}_\\d{2}-\\d{2}-\\d{2}(\\.\\w+)?$"
)

type FileConfig struct {
	FileDir       string
	FileName      string
	MaxFileSize   int64
	BackupCount   int
	BulkWriteSize int
	
	RotatorType   RotatorType
	Interval      int64 // unit: second. used in TimeRotator and TimeAndSizeRotator.
	TimeSuffixFmt string
	ReMatch       string
	FileSuffix    string
	
	MultiProcessWrite bool
	
	ErrCallback func(buf []byte, err error)
}

func NewFileConfig(opts ...FileOpt) *FileConfig {
	cfg := &FileConfig{
		MaxFileSize:   DefaultMaxFileSize,
		BackupCount:   DefaultBackupCount,
		BulkWriteSize: DefaultBulkWriteSize,
		Interval:      DefaultInterval,
		TimeSuffixFmt: FileRotatorSuffixFmt1,
		ReMatch:       FileRotatorReMatch1,
		FileSuffix:    DefaultFileSuffix,
	}
	
	for _, opt := range opts {
		opt(cfg)
	}
	
	return cfg
}

type FileHandler struct {
	formatter formatter.IFormatter
	filter    filter.IFilter
	rotator   IRotator
	
	bulkWriteSize int
	bufChan       chan []byte
	doneChan      chan bool
	done          bool
	
	ErrorCallback func(buf []byte, err error)
}

func NewFileHandler(handlerCfg *Config) (*FileHandler, error) {
	// Redirects the standard error output to the specified file,
	// in order to preserve the panic information during panic.
	// rewriteStderr(handlerCfg.File.FileDir, config.GlobalConfig.LoggerName)
	
	rotator, err := GetRotator4Config(handlerCfg.File)
	if err != nil {
		return nil, err
	}
	h := &FileHandler{
		formatter:     formatter.GetNewFormatter(handlerCfg.FormatConfig),
		filter:        handlerCfg.Filter,
		rotator:       rotator,
		bulkWriteSize: handlerCfg.File.BulkWriteSize,
		ErrorCallback: handlerCfg.File.ErrCallback,
	}
	h.init()
	return h, nil
}

func (h *FileHandler) init() {
	h.bufChan = make(chan []byte, 1024)
	h.doneChan = make(chan bool, 100)
	go h.flushWorker()
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
	select {
	case h.doneChan <- true: // send
		// default: // channel full
	}
	for i := 0; i < 100; i++ {
		if h.done {
			break
		}
		<-time.After(1 * time.Millisecond)
	}
	return h.rotator.Close()
}
