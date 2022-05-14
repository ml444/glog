package handlers

import (
	"errors"
	"fmt"
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/filters"
	"github.com/ml444/glog/formatters"
	"github.com/ml444/glog/message"
)

type IHandler interface {
	Emit(entry *message.Entry) error
	Flush()
	Sync() error
}

func GetNewHandler(handlerCfg config.BaseHandlerConfig) (IHandler, error) {
	formatter := formatters.GetNewFormatter(handlerCfg.Formatter)
	filter := filters.GetNewFilter(handlerCfg.Filter)

	switch handlerCfg.HandlerType {
	case config.HandlerTypeFile:
		return NewFileHandler(handlerCfg.File, formatter, filter)
	case config.HandlerTypeStream:
		return NewStreamHandler(formatter, filter)
	case config.HandlerTypeSyslog:
		return NewSyslogHandler(&handlerCfg.Syslog, formatter, filter)
	default:
		return NewSyslogHandler(&handlerCfg.Syslog, formatter, filter)
	}
}

type BaseHandler struct {
	formatter formatters.IFormatter
	filter    filters.IFilter
	//lock      sync.Mutex
}

//func (h *BaseHandler) Acquire() {
//	h.lock.Lock()
//}
//func (h *BaseHandler) Release() {
//	h.lock.Unlock()
//}

func (h *BaseHandler) Format(record *message.Entry) ([]byte, error) {
	if h.formatter == nil {
		return h.formatter.Format(record)
	}
	return nil, nil
}

func (h *BaseHandler) Handle(record *message.Entry) error {
	if h.filter != nil {
		if ok := h.filter.Filter(record); !ok {
			return errors.New(fmt.Sprintf("Filter out this msg: %v", record))
		}
	}

	msgByte, err := h.Format(record)
	if err != nil {
		return err
	}

	//h.Acquire()
	err = h.Emit(msgByte)
	//h.Release()
	return err
}

func (h *BaseHandler) Emit(msg []byte) error {
	fmt.Println("BaseHandler: ", string(msg))
	return nil
}
