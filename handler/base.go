package handler

import (
	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

type IHandler interface {
	Emit(entry *message.Entry) error
	Close() error
}

type HandlerType int

const (
	HandlerTypeStdout HandlerType = 0
	HandlerTypeFile   HandlerType = 1
	HandlerTypeStream HandlerType = 2
	HandlerTypeSyslog HandlerType = 3
)

type HandlerConfig struct {
	ExternalHandler IHandler
	HandlerType     HandlerType
	File            FileHandlerConfig
	Stream          StreamHandlerConfig
	Syslog          SyslogHandlerConfig

	Formatter formatter.FormatterConfig
	Filter    filter.IFilter
}

type StreamHandlerConfig struct {
	Streamer IStreamer
}
type SyslogHandlerConfig struct {
	Network  string
	Address  string
	Tag      string
	Priority int
}

func GetNewHandler(handlerCfg HandlerConfig) (IHandler, error) {
	if handlerCfg.ExternalHandler != nil {
		return handlerCfg.ExternalHandler, nil
	}
	switch handlerCfg.HandlerType {
	case HandlerTypeFile:
		return NewFileHandler(&handlerCfg)
	case HandlerTypeStream:
		return NewStreamHandler(&handlerCfg)
	case HandlerTypeSyslog:
		return NewSyslogHandler(&handlerCfg)
	default:
		return NewDefaultHandler(&handlerCfg)
	}
}
