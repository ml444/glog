package config

import (
	"os"
	"strings"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/level"
)

const (
	PatternTemplateWithDefault = "%[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s.%[Msecs]d %[LevelName]s %[Caller]s %[Message]v"
	PatternTemplateWithSimple  = "%[LevelName]s %[DateTime]s.%[Msecs]d %[Caller]s %[Message]v"
	PatternTemplateWithTrace   = "<%[TradeId]s> %[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s %[LevelName]s %[Caller]s %[Message]v"
)

func NewDefaultConfig() *Config {
	curDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}
	defaultLogDir := curDir
	defaultReportLogDir := curDir

	l := strings.Split(curDir, string(os.PathSeparator))
	defaultLogName := ""

	return &Config{
		LoggerName:      l[len(l)-1],
		LoggerLevel:     level.PrintLevel,
		LoggerCacheSize: 1024 * 64,

		EnableReport:    false,
		ReportLevel:     level.ErrorLevel,
		ReportCacheSize: 10000,

		ExitFunc: os.Exit,
		// TradeIDFunc:    nil,
		IsRecordCaller: true,
		LogHandlerConfig: handler.HandlerConfig{
			HandlerType: handler.HandlerTypeStdout,
			File: handler.FileHandlerConfig{
				RotatorType:       handler.FileRotatorTypeTimeAndSize,
				FileDir:           defaultLogDir,
				FileName:          "",
				MaxFileSize:       defaultMaxFileSize * 4,
				BulkWriteSize:     10485760, // 10MB
				BackupCount:       24,
				Interval:          60 * 60,
				TimeSuffixFmt:     "2006010215",
				ReMatch:           "^\\d{10}(\\.\\w+)?$",
				FileSuffix:        "log",
				MultiProcessWrite: false,

				ErrCallback: func(buf []byte, err error) {
					println("===>glog logger err: ", err.Error())
				},
			},
			Formatter: formatter.FormatterConfig{
				FormatterType:   formatter.FormatterTypeText,
				TimestampFormat: DefaultTimestampFormat,
				PatternStyle:    PatternTemplateWithDefault,
			},
		},
		ReportHandlerConfig: handler.HandlerConfig{
			HandlerType: handler.HandlerTypeFile,
			File: handler.FileHandlerConfig{
				RotatorType: handler.FileRotatorTypeSize,
				FileDir:     defaultReportLogDir,
				FileName:    defaultLogName,
				MaxFileSize: defaultMaxFileSize,
				BackupCount: 24,
				FileSuffix:  "report",

				ErrCallback: func(buf []byte, err error) {
					println("===>glog report err: ", err.Error())
				},
			},
			Formatter: formatter.FormatterConfig{
				FormatterType:   formatter.FormatterTypeJSON,
				TimestampFormat: DefaultTimestampFormat,
			},
		},
	}
}
