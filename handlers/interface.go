package handlers

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/filters"
	"github.com/ml444/glog/formatters"
	"github.com/ml444/glog/message"
)

type IHandler interface {
	//Emit(msg []byte) error

	//init(dir, name string) error
	Emit(entry *message.Entry) error
	Flush()
	Sync() error
}

func GetNewHandler(handlerCfg *config.BaseHandlerConfig) (IHandler, error) {
	formatter := formatters.GetNewFormatter(handlerCfg.Formatter)
	filter := filters.GetNewFilter(handlerCfg.Filter)

	switch handlerCfg.HandlerType {
	case config.HandlerTypeFile:
		return NewFileHandler(handlerCfg.File, formatter, filter)
	case config.HandlerTypeStream:
		return NewStreamHandler(formatter, filter)
	case config.HandlerTypeSyslog:
		return NewSyslogHandler(formatter, filter)
	default:
		return NewSyslogHandler(formatter, filter)
	}
}