package log

import (
	"os"

	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/message"
)

const (
	DefaultDateTimeFormat       = "01-02T15:04:05.000000" // microsecond
	defaultMaxFileSize    int64 = 1024 * 1024 * 1024
)

type Config struct {
	// Set the name of the Logger. If not set, the default is `glog`.
	LoggerName string

	// Set the global log level. Logs below this level will not be processed.
	// If not set, the default is `Info Level`.
	LoggerLevel Level

	// What level of logging is set here will trigger an exception to be thrown.
	// If this value is set, an exception will be thrown when an error of this level occurs.
	// You can only choose three levels: `FatalLevel`, `PanicLevel`, and `NoneLevel`.
	// The default setting is `NoneLevel`, which will not throw an exception.
	// For example, if it is set to FatalLevel, an exception will be thrown
	// when a FatalLevel error occurs.
	ThrowOnLevel Level

	// Disable recording of caller information
	DisableRecordCaller bool

	// For log processing configuration, multiple Worker coroutines can be set,
	// and each Worker can set different cache sizes, log levels, formatting
	// methods, filters, and output methods.
	// Output currently defines multiple modes, such as console, file, network, etc.
	// If not set, the default is a Worker, the cache size is 1000, the log level
	// is `Info Level`, and the output mode is the console.
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
}

func NewWorkerConfig(level Level, size int) *WorkerConfig {
	return &WorkerConfig{CacheSize: size, Level: level}
}

func (w *WorkerConfig) SetCacheSize(size int) *WorkerConfig {
	w.CacheSize = size
	return w
}

func (w *WorkerConfig) SetLevel(lvl Level) *WorkerConfig {
	w.Level = lvl
	return w
}

func (w *WorkerConfig) SetFileHandlerConfig(c *FileHandlerConfig) *WorkerConfig {
	w.HandlerCfg.File = c
	return w
}

func (w *WorkerConfig) SetStreamHandlerConfig(c *StreamHandlerConfig) *WorkerConfig {
	w.HandlerCfg.Stream = c
	return w
}

func (w *WorkerConfig) SetSyslogHandlerConfig(c *SyslogHandlerConfig) *WorkerConfig {
	w.HandlerCfg.Syslog = c
	return w
}

func (w *WorkerConfig) SetTextFormatterConfig(c *TextFormatterConfig) *WorkerConfig {
	w.FormatterCfg.Text = c
	return w
}

func (w *WorkerConfig) SetJSONFormatterConfig(c *JSONFormatterConfig) *WorkerConfig {
	w.FormatterCfg.JSON = c
	return w
}

func (w *WorkerConfig) SetXMLFormatterConfig(c *XMLFormatterConfig) *WorkerConfig {
	w.FormatterCfg.XML = c
	return w
}

func (w *WorkerConfig) SetHandler(h handler.IHandler) *WorkerConfig {
	w.CustomHandler = h
	return w
}

func (w *WorkerConfig) SetFormatter(f formatter.IFormatter) *WorkerConfig {
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
	//if c.ExitFunc == nil {
	//	c.ExitFunc = ExitHook
	//}
	if c.OnError == nil {
		c.OnError = func(_ interface{}, err error) {
			println("glog unexpected error: ", err.Error())
		}
	}

	if c.WorkerConfigList == nil {
		c.WorkerConfigList = []*WorkerConfig{
			NewDefaultStdoutWorkerConfig(),
		}
	} else {
		for i, workerCfg := range c.WorkerConfigList {
			if workerCfg.CacheSize == 0 {
				workerCfg.CacheSize = 1024
			}
			if workerCfg.Level == 0 {
				workerCfg.Level = PrintLevel
			}
			if workerCfg.CustomHandler != nil {
				continue
			}
			// remove Worker is nil
			if workerCfg == nil {
				c.WorkerConfigList = append(c.WorkerConfigList[:i], c.WorkerConfigList[i+1:]...)
			}
			if cc := workerCfg.FormatterCfg.Text; cc != nil && cc.TimeLayout == "" {
				cc.TimeLayout = DefaultDateTimeFormat
			}
			if cc := workerCfg.FormatterCfg.JSON; cc != nil && cc.TimeLayout == "" {
				cc.TimeLayout = DefaultDateTimeFormat
			}
			if cc := workerCfg.FormatterCfg.XML; cc != nil && cc.TimeLayout == "" {
				cc.TimeLayout = DefaultDateTimeFormat
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
	}
}
