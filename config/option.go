package config

import (
	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/level"
)

type OptionFunc func(config *Config)

// SetLoggerName Set the name of the logger, the default is the name of the program.
func SetLoggerName(name string) OptionFunc {
	return func(cfg *Config) {
		oldLoggerName := cfg.LoggerName
		cfg.LoggerName = name
		// If the file name is the same as the old logger name(default name), then the file name is also changed.
		if cfg.Handler.LogHandlerConfig.File.FileName == oldLoggerName {
			cfg.Handler.LogHandlerConfig.File.FileName = name
		}
		if cfg.Handler.ReportHandlerConfig.File.FileName == oldLoggerName {
			cfg.Handler.ReportHandlerConfig.File.FileName = name
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

func SetHandlerType2Logger(typ HandlerType) OptionFunc {
	return func(cfg *Config) { cfg.Handler.LogHandlerConfig.HandlerType = typ }
}
func SetHandlerType2Report(typ HandlerType) OptionFunc {
	return func(cfg *Config) { cfg.Handler.ReportHandlerConfig.HandlerType = typ }
}

// SetFileName2Logger By default, the file name is the same as the logger name,
// if you need to specify special can be set through this config.
func SetFileName2Logger(name string) OptionFunc {
	return func(cfg *Config) { cfg.Handler.LogHandlerConfig.File.FileName = name }
}

// SetFileName2Report By default, the file name is the same as the logger name,
// if you need to specify special can be set through this config.
func SetFileName2Report(name string) OptionFunc {
	return func(cfg *Config) { cfg.Handler.ReportHandlerConfig.File.FileName = name }
}

func SetFileDir2Logger(path string) OptionFunc {
	return func(cfg *Config) { cfg.Handler.LogHandlerConfig.File.FileDir = path }
}
func SetFileDir2Report(path string) OptionFunc {
	return func(cfg *Config) { cfg.Handler.ReportHandlerConfig.File.FileDir = path }
}

func SetFileRotatorType2Logger(typ RotatorType) OptionFunc {
	return func(cfg *Config) { cfg.Handler.LogHandlerConfig.File.RotatorType = typ }
}
func SetFileRotatorType2Report(typ RotatorType) OptionFunc {
	return func(cfg *Config) { cfg.Handler.ReportHandlerConfig.File.RotatorType = typ }
}

func SetFileBackupCount2Logger(count int) OptionFunc {
	return func(cfg *Config) { cfg.Handler.LogHandlerConfig.File.BackupCount = count }
}
func SetFileBackupCount2Report(count int) OptionFunc {
	return func(cfg *Config) { cfg.Handler.ReportHandlerConfig.File.BackupCount = count }
}

func SetFileMaxSize2Logger(size int64) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.LogHandlerConfig.File.MaxFileSize = size
	}
}
func SetFileMaxSize2Report(size int64) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.ReportHandlerConfig.File.MaxFileSize = size
	}
}

func SetFileWhen2Logger(when RotatorWhenType) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.LogHandlerConfig.File.When = when
	}
}
func SetFileWhen2Report(when RotatorWhenType) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.ReportHandlerConfig.File.When = when
	}
}

func SetFileRematch2Logger(pattern string) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.LogHandlerConfig.File.ReMatch = pattern
	}
}
func SetFileRematch2Report(pattern string) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.ReportHandlerConfig.File.ReMatch = pattern
	}
}

func SetFileTimeFmtSuffix2Logger(timeFmt string) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.LogHandlerConfig.File.TimeSuffixFmt = timeFmt
	}
}
func SetFileTimeFmtSuffix2Report(timeFmt string) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.ReportHandlerConfig.File.TimeSuffixFmt = timeFmt
	}
}

func SetFileHandlerConfig2Logger(fileCfg FileHandlerConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.LogHandlerConfig.File = fileCfg
	}
}
func SetFileHandlerConfig2Report(fileCfg FileHandlerConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.ReportHandlerConfig.File = fileCfg
	}
}

func SetStreamer2Logger(streamer filter.IStreamer) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.LogHandlerConfig.HandlerType = HandlerTypeStream
		cfg.Handler.LogHandlerConfig.Stream.Streamer = streamer
	}
}
func SetStreamer2Report(streamer filter.IStreamer) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.ReportHandlerConfig.HandlerType = HandlerTypeStream
		cfg.Handler.ReportHandlerConfig.Stream.Streamer = streamer
	}
}

func SetSyslogHandlerConfig2Logger(syslogCfg SyslogHandlerConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.LogHandlerConfig.Syslog = syslogCfg
	}
}
func SetSyslogHandlerConfig2Report(syslogCfg SyslogHandlerConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.ReportHandlerConfig.Syslog = syslogCfg
	}
}

func SetFormatterConfig2Logger(formatterCfg FormatterConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.LogHandlerConfig.Formatter = formatterCfg
	}
}
func SetFormatterConfig2Report(formatterCfg FormatterConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.ReportHandlerConfig.Formatter = formatterCfg
	}
}

func SetFormatterType2Logger(typ FormatterType) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.LogHandlerConfig.Formatter.FormatterType = typ
	}
}
func SetFormatterType2Report(typ FormatterType) OptionFunc {
	return func(cfg *Config) {
		cfg.Handler.ReportHandlerConfig.Formatter.FormatterType = typ
	}
}

// TODO: TO be continued
