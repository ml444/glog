package config

import (
	"time"
	
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
)

const (
	DefaultTimestampFormat       = time.RFC3339
	defaultMaxFileSize     int64 = 1024 * 1024 * 1024
)

type Config struct {
	*GeneralConfig
	*LogConfig
	*ReportConfig
}

type GeneralConfig struct {
	ExitOnFatal    bool
	ThrowOnPanic   bool
	IsRecordCaller bool
	EnableReport   bool
	
	ExitFunc    func(code int)
	TradeIDFunc func(entry *message.Entry) string
	OnError     func(msg *message.Entry, err error)
}

type BaseLogConfig struct {
	CacheSize int
	Level     level.LogLevel
	Config    *handler.Config
}

type LogConfig struct {
	*BaseLogConfig
	Name string
}

type ReportConfig struct {
	*BaseLogConfig
}
