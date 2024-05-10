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
	LoggerName          string
	LogHandlerConfig    handler.HandlerConfig
	ReportHandlerConfig handler.HandlerConfig
	ExitFunc            func(code int) // Function to exit the application, defaults to `os.Exit()`
	TradeIDFunc         func(entry *message.Entry) string
	OnError             func(msg *message.Entry, err error)
	ReportCacheSize     int
	LoggerCacheSize     int
	LoggerLevel         level.LogLevel
	ReportLevel         level.LogLevel
	ExitOnFatal         bool
	ThrowOnPanic        bool
	IsRecordCaller      bool
	EnableReport        bool
}