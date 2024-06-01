package config

const (
	defaultCacheSize         = 1024 * 64
	defaultFileBulkWriteSize = 24
	defaultReportFileSuffix  = "report"
	defaultFileTimeSuffixFmt = "2006010215"
	defaultFileReMatch       = "^\\d{10}(\\.\\w+)?$"
	defaultMaxFileSize       = 1024 * 1024 * 1024
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
