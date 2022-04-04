package config

import (
	"github.com/ml444/glog/levels"
	"time"
)

const (
	HandlerTypeFile   = 1
	HandlerTypeStream = 2
	HandlerTypeSyslog = 3

	FormatterTypeText = 1
	FormatterTypeJson = 2
	FormatterTypeXml  = 3

	FileRotatorTypeTime        = 1
	FileRotatorTypeSize        = 2
	FileRotatorTypeTimeAndSize = 3

	FileRotatorWhenSecond = 1
	FileRotatorWhenMinute = 2
	FileRotatorWhenHour   = 3
	FileRotatorWhenDay    = 4
)

const (
	DefaultTimestampFormat = time.RFC3339

	defaultMaxFileSize int64 = 4 * 1024 * 1024 * 1024

	defaultLogDir       string = "."
	defaultLogName       string = "glog"
	defaultReportLogDir string = "."
)

func NewDefaultConfig() *Config {
	return &Config{
		Logger: &LoggerConfig{
			Name:  "UNKNOWN",
			Level: levels.InfoLevel,
		},
		Engine: &EngineConfig{
			LogCacheSize: 100000,

			EnableReport:    false,
			ReportCacheSize: 10000,
			ReportLevel:     levels.WarnLevel,
		},
		Handler: &HandlerConfig{
			CommonConfig: &BaseHandlerConfig{
				HandlerType: HandlerTypeFile,
				File: &FileHandlerConfig{
					Rotator: &FileRotatorConfig{
						Type:         FileRotatorTypeTimeAndSize,
						FileDir:      defaultLogDir,
						FileName:     defaultLogName,
						MaxFileSize:  defaultMaxFileSize,
						When:         FileRotatorWhenHour,
						BackupCount:  50,
						IntervalStep: 1,
						SuffixFmt:    "2006010215",
						ReMatch:      "",
					},
				},
				Formatter: &FormatterConfig{
					FormatterType: FormatterTypeText,
					TimestampFormat: DefaultTimestampFormat,
					Text: &TextFormatterConfig{
						Pattern:                "%[ServiceName]s (%[Pid]d,%[RoutineId]d) %[Level]s %[FileName]s:%[CallerName]s:%[CallerLine]d %[Message]v",
						EnableQuote:            false,
						EnableQuoteEmptyFields: false,
						DisableColors:          false,
					},
				},
			},
			ReportConfig: &BaseHandlerConfig{
				HandlerType: HandlerTypeFile,
				File: &FileHandlerConfig{
					Rotator: &FileRotatorConfig{
						Type:        FileRotatorTypeSize,
						FileDir:     defaultReportLogDir,
						FileName:    defaultLogName,
						MaxFileSize: defaultMaxFileSize,
						BackupCount: 50,
					},
				},
				Formatter: &FormatterConfig{
					TimestampFormat: DefaultTimestampFormat,
					FormatterType: FormatterTypeJson,
					Json: &JSONFormatterConfig{},
				},
			},
		},

	}
}
