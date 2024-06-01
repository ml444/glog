package config

import (
	"os"
	"strings"
	
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/level"
)

const (
	defaultCacheSize         = 1024 * 64
	defaultFileBulkWriteSize = 24
	defaultReportFileSuffix  = "report"
	defaultFileTimeSuffixFmt = "2006010215"
	defaultFileReMatch       = "^\\d{10}(\\.\\w+)?$"
)

const (
	PatternTemplateWithDefault = "%[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s.%[Msecs]d %[LevelName]s %[Caller]s %[Message]v"
	PatternTemplateWithSimple  = "%[LevelName]s %[DateTime]s.%[Msecs]d %[Caller]s %[Message]v"
	PatternTemplateWithTrace   = "<%[TradeId]s> %[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s %[LevelName]s %[Caller]s %[Message]v"
)

var defaultFileErrCallback = func(buf []byte, err error) {
	if err != nil {
		println("===>glog logger err: ", err.Error())
	}
}

func NewDefaultGeneralConfig() *GeneralConfig {
	return &GeneralConfig{
		ExitOnFatal:    false,
		ThrowOnPanic:   false,
		IsRecordCaller: false,
		EnableReport:   false,
		ExitFunc:       os.Exit,
		TradeIDFunc:    nil,
		OnError:        nil,
	}
}

func NewDefaultLogConfig() *LogConfig {
	curDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}
	defaultLogDir := curDir
	l := strings.Split(curDir, string(os.PathSeparator))
	
	fileConfig := handler.NewFileConfig(
		handler.WithFileRotatorType(handler.FileRotatorTypeTimeAndSize),
		handler.WithFileDir(defaultLogDir),
		handler.WithFileMaxFileSize(defaultMaxFileSize),
		handler.WithFileBulkWriteSize(defaultFileBulkWriteSize),
		handler.WithFileTimeSuffixFmt(defaultFileTimeSuffixFmt),
		handler.WithFileReMatch(defaultFileReMatch),
		handler.WithFileErrCallback(defaultFileErrCallback),
	)
	
	formatConfig := formatter.NewConfig(
		formatter.WithFormatterType(formatter.TypeText),
		formatter.WithTimestampFormat(DefaultTimestampFormat),
		formatter.WithPatternStyle(PatternTemplateWithDefault),
	)
	
	handlerConfig := handler.NewConfig(
		handler.WithFileConfig(fileConfig),
		handler.WithType(handler.TypeStdout),
		handler.WithFormatConfig(formatConfig),
	)
	
	return &LogConfig{
		Name: l[len(l)-1],
		BaseLogConfig: &BaseLogConfig{
			CacheSize: defaultCacheSize,
			Level:     level.PrintLevel,
			Config:    handlerConfig,
		},
	}
}

func NewDefaultReportConfig() *ReportConfig {
	curDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}
	defaultReportLogDir := curDir
	
	fileConfig := handler.NewFileConfig(
		handler.WithFileRotatorType(handler.FileRotatorTypeSize),
		handler.WithFileDir(defaultReportLogDir),
		handler.WithFileMaxFileSize(defaultMaxFileSize),
		handler.WithFileBulkWriteSize(defaultFileBulkWriteSize),
		handler.WithFileFileSuffix(defaultReportFileSuffix),
		handler.WithFileErrCallback(defaultFileErrCallback),
	)
	
	formatConfig := formatter.NewConfig(
		formatter.WithFormatterType(formatter.TypeJson),
		formatter.WithTimestampFormat(DefaultTimestampFormat),
	)
	
	handlerConfig := handler.NewConfig(
		handler.WithFileConfig(fileConfig),
		handler.WithFormatConfig(formatConfig),
	)
	
	return &ReportConfig{
		BaseLogConfig: &BaseLogConfig{
			CacheSize: defaultCacheSize,
			Level:     level.PrintLevel,
			Config:    handlerConfig,
		},
	}
}

func NewDefaultConfig() *Config {
	return &Config{
		GeneralConfig: NewDefaultGeneralConfig(),
		LogConfig:     NewDefaultLogConfig(),
		ReportConfig:  NewDefaultReportConfig(),
	}
}
