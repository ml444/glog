package config

import (
	"os"
	"strings"
	"time"

	"github.com/ml444/glog/levels"
)

const (
	HandlerTypeDefault = 0
	HandlerTypeFile    = 1
	HandlerTypeStream  = 2
	HandlerTypeSyslog  = 3

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

	FileRotatorSuffixFmt1 = "20060102150405"
	FileRotatorSuffixFmt2 = "2006-01-02T15-04-05"
	FileRotatorSuffixFmt3 = "2006-01-02_15-04-05"

	FileRotatorReMatch1 = "^\\d{14}(\\.\\w+)?$"
	FileRotatorReMatch2 = "^\\d{4}-\\d{2}-\\d{2}T\\d{2}-\\d{2}-\\d{2}(\\.\\w+)?$"
	FileRotatorReMatch3 = "^\\d{4}-\\d{2}-\\d{2}_\\d{2}-\\d{2}-\\d{2}(\\.\\w+)?$"
)

const (
	DefaultTimestampFormat       = time.RFC3339
	defaultMaxFileSize     int64 = 1024 * 1024 * 1024
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
		LoggerName:      defaultLogName,
		LoggerLevel:     levels.InfoLevel,
		LoggerCacheSize: 100000,

		EnableReport:    false,
		ReportLevel:     levels.ErrorLevel,
		ReportCacheSize: 10000,
		IsRecordCaller:  true,
		Handler: HandlerConfig{
			LogHandlerConfig: BaseHandlerConfig{
				HandlerType: HandlerTypeDefault,
				File: FileHandlerConfig{
					Type:              FileRotatorTypeTimeAndSize,
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
						Pattern:                "%[LogName]s (%[Pid]d,%[RoutineId]d) %[Level]s %[FileName]s:%[CallerName]s:%[CallerLine]d %[Message]v",
						EnableQuote:            false,
						EnableQuoteEmptyFields: false,
						DisableColors:          false,
					},
				},
			},
			ReportHandlerConfig: BaseHandlerConfig{
				HandlerType: HandlerTypeFile,
				File: FileHandlerConfig{
					Type:        FileRotatorTypeSize,
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
