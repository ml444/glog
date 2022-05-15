package config

import "github.com/ml444/glog/levels"

type Config struct {
	LoggerName      string
	LoggerLevel     levels.LogLevel
	LoggerCacheSize int

	EnableReport    bool
	ReportLevel     levels.LogLevel
	ReportCacheSize int

	IsRecordCaller bool

	//Logger  LoggerConfig  `json:"logger"`
	//Engine  EngineConfig  `json:"engine"`
	Handler HandlerConfig `json:"handler"`
}

//type LoggerConfig struct {
//	Name           string
//	Level          levels.LogLevel
//	IsRecordCaller bool
//}
//
//type EngineConfig struct {
//	LogCacheSize int
//
//	EnableReport    bool
//	ReportCacheSize int
//	ReportLevel     levels.LogLevel
//}

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
	When         uint8
	IntervalStep int64
	SuffixFmt    string
	ReMatch      string


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
