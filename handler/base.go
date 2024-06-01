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

type Type int

const (
	TypeStdout Type = 0
	TypeFile   Type = 1
	TypeStream Type = 2
	TypeSyslog Type = 3
)

type Config struct {
	ExternalHandler IHandler
	
	Type   Type
	File   *FileConfig
	Stream *StreamConfig
	Syslog *SyslogConfig
	
	FormatConfig *formatter.Config
	Filter       filter.IFilter
}

func NewConfig(opts ...Opt) *Config {
	cfg := &Config{}
	// todo 是否需要默认值
	
	for _, opt := range opts {
		opt(cfg)
	}
	
	return cfg
}

type StreamConfig struct {
	Streamer IStreamer
}
type SyslogConfig struct {
	Network  string
	Address  string
	Tag      string
	Priority int
}

func GetNewHandler(handlerCfg *Config) (IHandler, error) {
	if handlerCfg.ExternalHandler != nil {
		return handlerCfg.ExternalHandler, nil
	}
	switch handlerCfg.Type {
	case TypeFile:
		return NewFileHandler(handlerCfg)
	case TypeStream:
		return NewStreamHandler(handlerCfg)
	case TypeSyslog:
		return NewSyslogHandler(handlerCfg)
	default:
		return NewDefaultHandler(handlerCfg)
	}
}
