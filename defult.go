package log

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/handler"
)

const (
	PatternTemplateWithDefault = "%[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s %[LevelName]s %[ShortCaller]s %[Message]v"
	PatternTemplateWithSimple  = "%[LevelName]s %[DateTime]s %[ShortCaller]s %[Message]v"
	PatternTemplateWithTrace   = "<%[TradeId]s> %[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s %[LevelName]s %[ShortCaller]s %[Message]v"
)

type RotatorType = handler.RotatorType

const (
	FileRotatorTypeTime        RotatorType = 1
	FileRotatorTypeSize        RotatorType = 2
	FileRotatorTypeTimeAndSize RotatorType = 3
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

func NewDefaultConfig() *Config {
	name := "glog"
	curDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
	} else {
		_, name = filepath.Split(curDir)
	}

	l := strings.Split(curDir, string(os.PathSeparator))
	if len(l) > 0 {
		name = l[len(l)-1]
	}

	return &Config{
		LoggerName:          name,
		LoggerLevel:         PrintLevel,
		ThrowOnLevel:        NoneLevel,
		ExitFunc:            ExitHook,
		WorkerConfigList:    []*WorkerConfig{NewDefaultStdoutHandlerConfig()},
		DisableRecordCaller: false,
	}
}

func NewDefaultStdoutHandlerConfig() *WorkerConfig {
	return &WorkerConfig{
		CacheSize:  1024,
		Level:      PrintLevel,
		HandlerCfg: HandlerConfig{},
		FormatterCfg: FormatterConfig{
			Text: &TextFormatterConfig{
				BaseFormatterConfig: BaseFormatterConfig{
					TimeLayout:  DefaultDateTimeFormat,
					EnableColor: true,
					ShortLevel:  true,
				},
				PatternStyle:           PatternTemplateWithDefault,
				EnableQuote:            false,
				EnableQuoteEmptyFields: false,
			},
		},
	}
}

func NewDefaultFileTextHandlerConfig() *WorkerConfig {
	curDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}
	return &WorkerConfig{
		CacheSize: 1024 * 64,
		Level:     PrintLevel,
		HandlerCfg: HandlerConfig{
			File: &FileHandlerConfig{
				RotatorType:       FileRotatorTypeTimeAndSize,
				FileDir:           curDir,
				FileName:          "",
				MaxFileSize:       defaultMaxFileSize * 4,
				BulkWriteSize:     10485760, // 10MB
				BackupCount:       24,
				Interval:          60 * 60,
				TimeSuffixFmt:     "2006010215",
				ReMatch:           "^\\d{10}(\\.\\w+)?$",
				FileSuffix:        "log",
				MultiProcessWrite: false,
			},
		},
		FormatterCfg: FormatterConfig{
			Text: &TextFormatterConfig{
				BaseFormatterConfig: BaseFormatterConfig{
					TimeLayout:  DefaultDateTimeFormat,
					EnableColor: false,
				},
				PatternStyle:           PatternTemplateWithDefault,
				EnableQuote:            false,
				EnableQuoteEmptyFields: false,
			},
		},
	}
}

func NewDefaultFileJsonHandlerConfig() WorkerConfig {
	curDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}
	return WorkerConfig{
		CacheSize: 1000,
		Level:     ErrorLevel,
		HandlerCfg: HandlerConfig{
			File: &FileHandlerConfig{
				RotatorType: FileRotatorTypeSize,
				FileDir:     curDir,
				FileName:    "",
				MaxFileSize: defaultMaxFileSize,
				BackupCount: 24,
				FileSuffix:  "report",
			},
		},
		FormatterCfg: FormatterConfig{
			JSON: &JSONFormatterConfig{
				BaseFormatterConfig: BaseFormatterConfig{
					TimeLayout: DefaultDateTimeFormat,
				},
				PrettyPrint: true,
			},
		},
	}
}

func newHandler(workerCfg WorkerConfig) (handler.IHandler, error) {
	if workerCfg.CustomHandler != nil {
		return workerCfg.CustomHandler, nil
	}
	fm := workerCfg.CustomFormatter
	if fm == nil {
		fm = newFormatter(workerCfg.FormatterCfg)
	}
	handlerCfg := workerCfg.HandlerCfg
	if handlerCfg.File != nil {
		return handler.NewFileHandler(handlerCfg.File, fm, workerCfg.CustomFilter)
	}
	if handlerCfg.Stream != nil {
		return handler.NewStreamHandler(handlerCfg.Stream, fm, workerCfg.CustomFilter)
	}
	if handlerCfg.Syslog != nil {
		return handler.NewSyslogHandler(handlerCfg.Syslog, fm, workerCfg.CustomFilter)
	}
	return handler.NewStdoutHandler(fm, workerCfg.CustomFilter)
}

func newFormatter(formatterCfg FormatterConfig) formatter.IFormatter {
	if formatterCfg.Text != nil {
		return formatter.NewTextFormatter(*formatterCfg.Text)
	}
	if formatterCfg.JSON != nil {
		return formatter.NewJSONFormatter(*formatterCfg.JSON)
	}
	if formatterCfg.XML != nil {
		return formatter.NewXMLFormatter(*formatterCfg.XML)
	}
	return formatter.NewTextFormatter(TextFormatterConfig{
		BaseFormatterConfig: BaseFormatterConfig{
			TimeLayout: DefaultDateTimeFormat,
		},
		PatternStyle: PatternTemplateWithDefault,
	})
}
