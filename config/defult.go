package config

import (
	"github.com/ml444/glog/levels"
	"github.com/ml444/glog/message"
	"os"
	"strings"
)

var (
	defaultLogDir       string
	defaultReportLogDir string
	defaultLogName      string
)

func init() {
	curDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}
	defaultLogDir = curDir
	defaultReportLogDir = curDir

	l := strings.Split(curDir, string(os.PathSeparator))
	defaultLogName = l[len(l)-1]
}

func NewDefaultConfig() *Config {
	return &Config{
		EngineType: EngineTypeChannel,

		LoggerName:      defaultLogName,
		LoggerLevel:     levels.InfoLevel,
		LoggerCacheSize: 1024 * 64,

		EnableReport:    false,
		ReportLevel:     levels.ErrorLevel,
		ReportCacheSize: 10000,

		ExitFunc: os.Exit,
		TradeIDFunc: func(entry *message.Entry) string {
			return "TradeId"
		},
		IsRecordCaller: true,
		Handler: HandlerConfig{
			LogHandlerConfig: BaseHandlerConfig{
				HandlerType: HandlerTypeDefault,
				File: FileHandlerConfig{
					RotatorType:       FileRotatorTypeTimeAndSize,
					FileDir:           defaultLogDir,
					FileName:          defaultLogName,
					MaxFileSize:       defaultMaxFileSize * 4,
					When:              FileRotatorWhenHour,
					BackupCount:       50,
					IntervalStep:      1,
					TimeSuffixFmt:     "2006010215",
					ReMatch:           "^\\d{10}(\\.\\w+)?$",
					FileSuffix:        "log",
					MultiProcessWrite: false,

					ErrCallback: func(err error) {
						println("===> logger err: ", err)
					},
				},
				Formatter: FormatterConfig{
					FormatterType:   FormatterTypeText,
					TimestampFormat: DefaultTimestampFormat,
					Text: TextFormatterConfig{
						Pattern:                "<%[TradeId]s> %[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s %[LevelName]s %[Caller]s %[Message]v",
						EnableQuote:            false,
						EnableQuoteEmptyFields: false,
						DisableColors:          false,
					},
				},
			},
			ReportHandlerConfig: BaseHandlerConfig{
				HandlerType: HandlerTypeFile,
				File: FileHandlerConfig{
					RotatorType: FileRotatorTypeSize,
					FileDir:     defaultReportLogDir,
					FileName:    defaultLogName,
					MaxFileSize: defaultMaxFileSize,
					BackupCount: 50,
					FileSuffix:  "report",

					ErrCallback: func(err error) {
						println("===> report err: ", err)
					},
				},
				Formatter: FormatterConfig{
					TimestampFormat: DefaultTimestampFormat,
					FormatterType:   FormatterTypeJson,
					Json:            JSONFormatterConfig{},
				},
			},
		},
	}
}

var GlobalConfig = NewDefaultConfig()
