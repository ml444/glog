package handler

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/message"
)

type IHandler interface {
	Emit(entry *message.Entry) error
	Close() error
}

func GetNewHandler(handlerCfg config.BaseHandlerConfig) (IHandler, error) {

	switch handlerCfg.HandlerType {
	case config.HandlerTypeFile:
		return NewFileHandler(&handlerCfg)
	case config.HandlerTypeStream:
		return NewStreamHandler(&handlerCfg)
	case config.HandlerTypeSyslog:
		return NewSyslogHandler(&handlerCfg)
	default:
		return NewDefaultHandler(&handlerCfg)
	}
}
