package config

import (
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/level"
)

type OptionFunc func(config *Config)

// SetLoggerName Set the name of the logger, the default is the name of the program.
func SetLoggerName(name string) OptionFunc {
	return func(cfg *Config) {
		oldLoggerName := cfg.LoggerName
		cfg.LoggerName = name
		// If the file name is the same as the old logger name(default name), then the file name is also changed.
		if cfg.LogHandlerConfig.File.FileName == oldLoggerName {
			cfg.LogHandlerConfig.File.FileName = name
		}
		if cfg.ReportHandlerConfig.File.FileName == oldLoggerName {
			cfg.ReportHandlerConfig.File.FileName = name
		}
	}
}

func EnableExitOnFatal() OptionFunc {
	return func(cfg *Config) { cfg.ExitOnFatal = true }
}

func SetLevel2Logger(lvl level.LogLevel) OptionFunc {
	return func(cfg *Config) { cfg.LoggerLevel = lvl }
}

func SetLevel2Report(lvl level.LogLevel) OptionFunc {
	return func(cfg *Config) { cfg.ReportLevel = lvl }
}

func SetEnableReport() OptionFunc {
	return func(cfg *Config) { cfg.EnableReport = true }
}

func SetCacheSize2Logger(size int) OptionFunc {
	return func(cfg *Config) { cfg.LoggerCacheSize = size }
}

func SetCacheSize2Report(size int) OptionFunc {
	return func(cfg *Config) { cfg.ReportCacheSize = size }
}

func SetHandlerType2Logger(typ handler.HandlerType) OptionFunc {
	return func(cfg *Config) { cfg.LogHandlerConfig.HandlerType = typ }
}

func SetHandlerType2Report(typ handler.HandlerType) OptionFunc {
	return func(cfg *Config) { cfg.ReportHandlerConfig.HandlerType = typ }
}

// SetFileName2Logger By default, the file name is the same as the logger name,
// if you need to specify special can be set through this config.
func SetFileName2Logger(name string) OptionFunc {
	return func(cfg *Config) { cfg.LogHandlerConfig.File.FileName = name }
}

// SetFileName2Report By default, the file name is the same as the logger name,
// if you need to specify special can be set through this config.
func SetFileName2Report(name string) OptionFunc {
	return func(cfg *Config) { cfg.ReportHandlerConfig.File.FileName = name }
}

func SetFileDir2Logger(path string) OptionFunc {
	return func(cfg *Config) { cfg.LogHandlerConfig.File.FileDir = path }
}

func SetFileDir2Report(path string) OptionFunc {
	return func(cfg *Config) { cfg.ReportHandlerConfig.File.FileDir = path }
}

func SetFileRotatorType2Logger(typ handler.RotatorType) OptionFunc {
	return func(cfg *Config) { cfg.LogHandlerConfig.File.RotatorType = typ }
}

func SetFileRotatorType2Report(typ handler.RotatorType) OptionFunc {
	return func(cfg *Config) { cfg.ReportHandlerConfig.File.RotatorType = typ }
}

func SetFileBackupCount2Logger(count int) OptionFunc {
	return func(cfg *Config) { cfg.LogHandlerConfig.File.BackupCount = count }
}

func SetFileBackupCount2Report(count int) OptionFunc {
	return func(cfg *Config) { cfg.ReportHandlerConfig.File.BackupCount = count }
}

func SetFileMaxSize2Logger(size int64) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.File.MaxFileSize = size
	}
}

func SetFileMaxSize2Report(size int64) OptionFunc {
	return func(cfg *Config) {
		cfg.ReportHandlerConfig.File.MaxFileSize = size
	}
}

func SetFileRolloverInterval2Logger(interval int64) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.File.Interval = interval
	}
}

func SetFileRolloverInterval2Report(interval int64) OptionFunc {
	return func(cfg *Config) {
		cfg.ReportHandlerConfig.File.Interval = interval
	}
}

func SetFileRematch2Logger(pattern string) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.File.ReMatch = pattern
	}
}

func SetFileRematch2Report(pattern string) OptionFunc {
	return func(cfg *Config) {
		cfg.ReportHandlerConfig.File.ReMatch = pattern
	}
}

func SetFileTimeFmtSuffix2Logger(timeFmt string) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.File.TimeSuffixFmt = timeFmt
	}
}

func SetFileTimeFmtSuffix2Report(timeFmt string) OptionFunc {
	return func(cfg *Config) {
		cfg.ReportHandlerConfig.File.TimeSuffixFmt = timeFmt
	}
}

func SetFileHandlerConfig2Logger(fileCfg handler.FileHandlerConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.File = fileCfg
	}
}

func SetFileHandlerConfig2Report(fileCfg handler.FileHandlerConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.ReportHandlerConfig.File = fileCfg
	}
}

func SetStreamer2Logger(streamer handler.IStreamer) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.HandlerType = handler.HandlerTypeStream
		cfg.LogHandlerConfig.Stream.Streamer = streamer
	}
}

func SetStreamer2Report(streamer handler.IStreamer) OptionFunc {
	return func(cfg *Config) {
		cfg.ReportHandlerConfig.HandlerType = handler.HandlerTypeStream
		cfg.ReportHandlerConfig.Stream.Streamer = streamer
	}
}

func SetSyslogHandlerConfig2Logger(syslogCfg handler.SyslogHandlerConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.Syslog = syslogCfg
	}
}

func SetSyslogHandlerConfig2Report(syslogCfg handler.SyslogHandlerConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.ReportHandlerConfig.Syslog = syslogCfg
	}
}

func SetFormatterConfig2Logger(formatterCfg formatter.FormatterConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.Formatter = formatterCfg
	}
}

func SetFormatterConfig2Report(formatterCfg formatter.FormatterConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.ReportHandlerConfig.Formatter = formatterCfg
	}
}

func SetFormatterType2Logger(typ formatter.FormatterType) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.Formatter.FormatterType = typ
	}
}

func SetFormatterType2Report(typ formatter.FormatterType) OptionFunc {
	return func(cfg *Config) {
		cfg.ReportHandlerConfig.Formatter.FormatterType = typ
	}
}
