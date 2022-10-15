package config

import (
	"github.com/ml444/glog/levels"
	"github.com/ml444/glog/message"
)

type Config struct {
	ExitOnFatal    bool
	ExitOnPanic    bool
	IsRecordCaller bool
	EnableReport   bool

	LoggerLevel     levels.LogLevel
	ReportLevel     levels.LogLevel
	ReportCacheSize int
	LoggerCacheSize int
	EngineType      EngineType

	LoggerName string

	Handler HandlerConfig `json:"handler"`

	TradeIDFunc func(entry *message.Entry) string
	OnError     func(msg *message.Entry, err error)
}

type HandlerConfig struct {
	LogHandlerConfig    BaseHandlerConfig
	ReportHandlerConfig BaseHandlerConfig
}

type BaseHandlerConfig struct {
	HandlerType uint8
	File        FileHandlerConfig
	Stream      StreamHandlerConfig
	Syslog      SyslogHandlerConfig

	Formatter FormatterConfig
	Filter    FilterConfig
}

type FileHandlerConfig struct {
	Type        int8
	FileDir     string
	FileName    string
	MaxFileSize int64
	BackupCount int

	// TimeRotator and TimeAndSizeRotator
	When          uint8
	IntervalStep  int64
	TimeSuffixFmt string
	ReMatch       string
	FileSuffix    string

	MultiProcessWrite bool

	ErrCallback func(err error)
}

type StreamHandlerConfig struct {
}
type SyslogHandlerConfig struct {
	Network  string
	Address  string
	Priority int
	Tag      string
}
type FormatterConfig struct {
	TimestampFormat string
	FormatterType   uint8
	Text            TextFormatterConfig
	Json            JSONFormatterConfig
	Xml             XMLFormatterConfig
}
type TextFormatterConfig struct {
	Pattern                string
	EnableQuote            bool
	EnableQuoteEmptyFields bool
	DisableColors          bool
}
type JSONFormatterConfig struct {
}

type XMLFormatterConfig struct {
}
type FilterConfig struct {
}
