package config

import (
	"github.com/ml444/glog/level"
	"os"
	"strings"
)

const (
	PatternTemplate1 = "%[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s.%[Msecs]d %[LevelName]s %[Caller]s %[Message]v"
	PatternTemplate2 = "<%[TradeId]s> %[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s %[LevelName]s %[Caller]s %[Message]v"
)

func NewDefaultConfig() *Config {
	curDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}
	defaultLogDir := curDir
	defaultReportLogDir := curDir

	l := strings.Split(curDir, string(os.PathSeparator))
	defaultLogName := l[len(l)-1]

	return &Config{

		LoggerName:      defaultLogName,
		LoggerLevel:     level.PrintLevel,
		LoggerCacheSize: 1024 * 64,

		EnableReport:    false,
		ReportLevel:     level.ErrorLevel,
		ReportCacheSize: 10000,

		ExitFunc: os.Exit,
		//TradeIDFunc:    nil,
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
					BackupCount:       24,
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
						PatternStyle:           PatternTemplate1,
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
					BackupCount: 24,
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
