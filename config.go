package log

import (
	"os"

	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/message"
)

const (
	DefaultDateTimeFormat       = formatter.DefaultTimeLayout // microsecond
	defaultMaxFileSize    int64 = 1024 * 1024 * 1024
)

type Config struct {
	// Set the name of the Logger. If not set, the default is `glog`.
	LoggerName string

	// Set the global log level. Logs below this level will not be processed.
	// NewDefaultConfig uses PrintLevel. A zero value Config is normalized to InfoLevel.
	LoggerLevel Level

	// What level of logging is set here will trigger an exception to be thrown.
	// If this value is set, an exception will be thrown when an error of this level occurs.
	// You can only choose three levels: `FatalLevel`, `PanicLevel`, and `NoneLevel`.
	// The default setting is `NoneLevel`, which will not throw an exception.
	// For example, if it is set to FatalLevel, an exception will be thrown
	// when a FatalLevel error occurs.
	ThrowOnLevel Level

	// Enable recording of caller information
	EnableRecordCaller bool

	CallerSkipCount int

	// EnableRoutineID controls whether each entry records its goroutine ID.
	// It defaults to true for backward compatibility. Disable it on hot paths
	// when the ID is not part of the desired output.
	EnableRoutineID *bool

	// enable color rendering. Only enabled by default in the text formatter.
	EnableColorRender *bool

	// time layout string, for example: "2006-01-02 15:04:05.000"
	TimeLayout string

	// For log processing configuration, multiple Worker coroutines can be set,
	// and each Worker can set different cache sizes, log levels, formatting
	// methods, filters, and output methods.
	// Output currently defines multiple modes, such as console, file, network, etc.
	// If not set, the default is one Worker with a cache size of 1024, following
	// the logger level and writing to the console.
	// Of course, if there is no suitable output method, you can customize the Handler.
	WorkerConfigList []*WorkerConfig

	// When ThrowOnLevel is set, if a ThrowOnLevel error occurs, this function
	// will be called before exiting. The default is `os.Exit()`.
	ExitFunc func(code int)

	// If an unexpected error occurs in asynchronous logic, you can use this
	// callback function to handle the error, such as saving it to other places
	// or alerting notifications, etc. The default is to print to stderr.
	OnError func(v interface{}, err error)

	// Link tracking solutions are often used in microservices, where the
	// tracking ID is the core of the entire call link. Customize this
	// function to return the Trace ID, and then record it in the log.
	TraceIDFunc func(entry *message.Entry) string
}

type FileHandlerConfig = handler.FileHandlerConfig
type StreamHandlerConfig = handler.StreamHandlerConfig
type SyslogHandlerConfig = handler.SyslogHandlerConfig
type BackpressureStrategy = handler.BackpressureStrategy
type BackpressureConfig = handler.BackpressureConfig
type BackpressureStats = handler.BackpressureStats

type BaseFormatterConfig = formatter.BaseFormatterConfig
type TextFormatterConfig = formatter.TextFormatterConfig
type JSONFormatterConfig = formatter.JSONFormatterConfig
type XMLFormatterConfig = formatter.XMLFormatterConfig

type HandlerConfig struct {
	File   *FileHandlerConfig
	Stream *StreamHandlerConfig
	Syslog *SyslogHandlerConfig
}

type FormatterConfig struct {
	Text *TextFormatterConfig
	JSON *JSONFormatterConfig
	XML  *XMLFormatterConfig
}

type WorkerConfig struct {
	CacheSize       int
	Level           Level
	HandlerCfg      HandlerConfig
	FormatterCfg    FormatterConfig
	CustomHandler   handler.IHandler
	CustomFilter    filter.IFilter
	CustomFormatter formatter.IFormatter
	Backpressure    BackpressureConfig
	loggerName      string
}

func NewWorkerConfig(level Level, size int) *WorkerConfig {
	return &WorkerConfig{CacheSize: size, Level: level}
}

func (w *WorkerConfig) SetCacheSize(size int) *WorkerConfig {
	w.CacheSize = size
	return w
}

func (w *WorkerConfig) SetBackpressure(config BackpressureConfig) *WorkerConfig {
	w.Backpressure = config
	return w
}

func (w *WorkerConfig) SetBackpressureStrategy(strategy BackpressureStrategy) *WorkerConfig {
	w.Backpressure.Strategy = strategy
	return w
}

func (w *WorkerConfig) SetLevel(lvl Level) *WorkerConfig {
	w.Level = lvl
	return w
}

func (w *WorkerConfig) SetFileHandlerConfig(c *FileHandlerConfig) *WorkerConfig {
	w.HandlerCfg = HandlerConfig{File: c}
	w.CustomHandler = nil
	return w
}

func (w *WorkerConfig) SetStreamHandlerConfig(c *StreamHandlerConfig) *WorkerConfig {
	w.HandlerCfg = HandlerConfig{Stream: c}
	w.CustomHandler = nil
	return w
}

func (w *WorkerConfig) SetSyslogHandlerConfig(c *SyslogHandlerConfig) *WorkerConfig {
	w.HandlerCfg = HandlerConfig{Syslog: c}
	w.CustomHandler = nil
	return w
}

func (w *WorkerConfig) SetTextFormatterConfig(c *TextFormatterConfig) *WorkerConfig {
	w.FormatterCfg = FormatterConfig{Text: c}
	w.CustomFormatter = nil
	return w
}

func (w *WorkerConfig) SetJSONFormatterConfig(c *JSONFormatterConfig) *WorkerConfig {
	w.FormatterCfg = FormatterConfig{JSON: c}
	w.CustomFormatter = nil
	return w
}

func (w *WorkerConfig) SetXMLFormatterConfig(c *XMLFormatterConfig) *WorkerConfig {
	w.FormatterCfg = FormatterConfig{XML: c}
	w.CustomFormatter = nil
	return w
}

func (w *WorkerConfig) SetHandler(h handler.IHandler) *WorkerConfig {
	w.HandlerCfg = HandlerConfig{}
	w.CustomHandler = h
	return w
}

func (w *WorkerConfig) SetFormatter(f formatter.IFormatter) *WorkerConfig {
	w.FormatterCfg = FormatterConfig{}
	w.CustomFormatter = f
	return w
}

func (w *WorkerConfig) SetFilter(f filter.IFilter) *WorkerConfig {
	w.CustomFilter = f
	return w
}

func (c *Config) Check() {
	if c.LoggerLevel == 0 {
		c.LoggerLevel = InfoLevel
	}
	if c.ThrowOnLevel == 0 {
		c.ThrowOnLevel = NoneLevel
	}
	if c.TimeLayout == "" {
		c.TimeLayout = DefaultDateTimeFormat
	}
	if c.ExitFunc == nil {
		c.ExitFunc = os.Exit
	}
	if c.OnError == nil {
		c.OnError = func(_ interface{}, err error) {
			println("glog unexpected error: ", err.Error())
		}
	}

	if c.WorkerConfigList == nil {
		c.WorkerConfigList = []*WorkerConfig{
			NewDefaultStdoutWorkerConfig(),
		}
	}
	var validWorkerConfigs []*WorkerConfig
	for _, workerCfg := range c.WorkerConfigList {
		if workerCfg == nil {
			continue
		}
		validWorkerConfigs = append(validWorkerConfigs, workerCfg)
		workerCfg.loggerName = c.LoggerName
		if workerCfg.CacheSize == 0 {
			workerCfg.CacheSize = 1024
		}
		workerCfg.Backpressure = workerCfg.Backpressure.Normalize(handler.BackpressureStrategyBlock)
		if workerCfg.CustomHandler != nil {
			continue
		}
		if workerCfg.CustomFormatter == nil && workerCfg.FormatterCfg.Text == nil && workerCfg.FormatterCfg.JSON == nil && workerCfg.FormatterCfg.XML == nil {
			workerCfg.FormatterCfg.Text = &TextFormatterConfig{PatternStyle: PatternTemplateWithDefault}
		}
		if cc := workerCfg.FormatterCfg.Text; cc != nil {
			if cc.TimeLayout == "" {
				cc.TimeLayout = c.TimeLayout
			}
			if c.EnableColorRender != nil {
				cc.EnableColor = *c.EnableColorRender
			}
		}
		if cc := workerCfg.FormatterCfg.JSON; cc != nil {
			if cc.TimeLayout == "" {
				cc.TimeLayout = c.TimeLayout
			}
		}
		if cc := workerCfg.FormatterCfg.XML; cc != nil {
			if cc.TimeLayout == "" {
				cc.TimeLayout = c.TimeLayout
			}
		}
		if cc := workerCfg.HandlerCfg.File; cc != nil {
			if cc.FileName == "" {
				cc.FileName = c.LoggerName
			}
			if cc.FileDir == "" {
				curDir, err := os.Getwd()
				if err != nil {
					println(err.Error())
				} else {
					cc.FileDir = curDir
				}
			}
			if cc.RotatorType == 0 {
				cc.RotatorType = handler.FileRotatorTypeTimeAndSize
			}
			if cc.MaxFileSize == 0 {
				cc.MaxFileSize = defaultMaxFileSize
			}
			if cc.BulkWriteSize == 0 {
				cc.BulkWriteSize = 10485760
			}
			if cc.BufferSize == 0 {
				cc.BufferSize = 10000
			}
			cc.Backpressure = cc.Backpressure.Normalize(handler.BackpressureStrategyDrop)
			if cc.Interval == 0 {
				cc.Interval = 60 * 60
			}
			if cc.TimeSuffixFmt == "" {
				cc.TimeSuffixFmt = "2006010215"
			}
			if cc.ReMatch == "" {
				cc.ReMatch = `^\d{4}\d{2}\d{2}\d{2}(\.\w+)?$`
			}
			if cc.FileSuffix == "" {
				cc.FileSuffix = "log"
			}
			if cc.ErrCallback == nil {
				cc.ErrCallback = c.OnError
			}
		}

		if cc := workerCfg.HandlerCfg.Stream; cc != nil {
			if cc.Streamer == nil {
				cc.Streamer = os.Stdout
			}
		}
		if cc := workerCfg.HandlerCfg.Syslog; cc != nil {
			if cc.Network == "" {
				cc.Network = "udp"
			}
			if cc.Address == "" {
				cc.Address = "localhost:514"
			}
			if cc.Tag == "" {
				cc.Tag = c.LoggerName
			}
		}
	}
	c.WorkerConfigList = validWorkerConfigs
}
