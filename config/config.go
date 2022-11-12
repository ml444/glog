package config

import (
	"github.com/ml444/glog/inter"
	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
	"time"
)

type EngineType int

const (
	EngineTypeChannel    EngineType = 1
	EngineTypeRingBuffer EngineType = 2
)

type HandlerType int

const (
	HandlerTypeDefault HandlerType = 0
	HandlerTypeFile    HandlerType = 1
	HandlerTypeStream  HandlerType = 2
	HandlerTypeSyslog  HandlerType = 3
)

type FormatterType int

const (
	FormatterTypeText FormatterType = 1
	FormatterTypeJson FormatterType = 2
	FormatterTypeXml  FormatterType = 3
)

type RotatorType int

const (
	FileRotatorTypeTime        RotatorType = 1
	FileRotatorTypeSize        RotatorType = 2
	FileRotatorTypeTimeAndSize RotatorType = 3
)

type RotatorWhenType int

const (
	FileRotatorWhenSecond RotatorWhenType = 1
	FileRotatorWhenMinute RotatorWhenType = 2
	FileRotatorWhenHour   RotatorWhenType = 3
	FileRotatorWhenDay    RotatorWhenType = 4
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

const (
	DefaultTimestampFormat       = time.RFC3339
	defaultMaxFileSize     int64 = 1024 * 1024 * 1024
)

type Config struct {
	ExitOnFatal    bool
	IsRecordCaller bool
	EnableReport   bool

	LoggerLevel     level.LogLevel
	ReportLevel     level.LogLevel
	ReportCacheSize int
	LoggerCacheSize int
	EngineType      EngineType

	LoggerName string

	Handler HandlerConfig `json:"handler"`

	ExitFunc    func(code int) // Function to exit the application, defaults to `os.Exit()`
	TradeIDFunc func(entry *message.Entry) string
	OnError     func(msg *message.Entry, err error)
}

type HandlerConfig struct {
	LogHandlerConfig    BaseHandlerConfig
	ReportHandlerConfig BaseHandlerConfig
}

type BaseHandlerConfig struct {
	HandlerType HandlerType
	File        FileHandlerConfig
	Stream      StreamHandlerConfig
	Syslog      SyslogHandlerConfig

	Formatter FormatterConfig
	Filter    inter.IFilter
}

type FileHandlerConfig struct {
	RotatorType RotatorType
	FileDir     string
	FileName    string
	MaxFileSize int64
	BackupCount int

	When          RotatorWhenType // used in TimeRotator and TimeAndSizeRotator
	IntervalStep  int64
	TimeSuffixFmt string
	ReMatch       string
	FileSuffix    string

	MultiProcessWrite bool

	ErrCallback func(err error)
}

type StreamHandlerConfig struct {
	Streamer inter.IStreamer
}
type SyslogHandlerConfig struct {
	Network  string
	Address  string
	Priority int
	Tag      string
}
type FormatterConfig struct {
	TimestampFormat string
	FormatterType   FormatterType
	Text            TextFormatterConfig
	Json            JSONFormatterConfig
	Xml             XMLFormatterConfig
}
type TextFormatterConfig struct {
	PatternStyle           string // style template for formatting the data, which determines the order of the fields and the presentation style.
	EnableQuote            bool   // keep the string literal, while escaping safely if necessary.
	EnableQuoteEmptyFields bool   // when the value of field is empty, keep the string literal.
	DisableColors          bool   // adding color rendering to the output.
}
type JSONFormatterConfig struct {
	DisableTimestamp  bool // allows disabling automatic timestamps in output.
	DisableHTMLEscape bool // allows disabling html escaping in output.
	PrettyPrint       bool // will indent all json logs.
}

type XMLFormatterConfig struct {
}
