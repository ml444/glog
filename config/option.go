package config

import "github.com/ml444/glog/level"

type OptionFunc func(config *Config)

func SetLoggerName(name string) OptionFunc {
	return func(cfg *Config) { cfg.LoggerName = name }
}

func SetLevel2Logger(lvl level.LogLevel) OptionFunc {
	return func(cfg *Config) { cfg.LoggerLevel = lvl }
}
func SetLevel2Report(lvl level.LogLevel) OptionFunc {
	return func(cfg *Config) { cfg.ReportLevel = lvl }
}

func SetEnableReport(enable bool) OptionFunc {
	return func(cfg *Config) { cfg.EnableReport = enable }
}

func SetCacheSize2Logger(size int) OptionFunc {
	return func(cfg *Config) { cfg.LoggerCacheSize = size }
}
func SetCacheSize2Report(size int) OptionFunc {
	return func(cfg *Config) { cfg.ReportCacheSize = size }
}

func SetFileName2Logger(name string) OptionFunc {
	return func(cfg *Config) { cfg.Handler.LogHandlerConfig.File.FileName = name }
}
func SetFileName2Report(name string) OptionFunc {
	return func(cfg *Config) { cfg.Handler.ReportHandlerConfig.File.FileName = name }
}

func SetFileDir2Logger(path string) OptionFunc {
	return func(cfg *Config) { cfg.Handler.LogHandlerConfig.File.FileDir = path }
}
func SetFileDir2Report(path string) OptionFunc {
	return func(cfg *Config) { cfg.Handler.ReportHandlerConfig.File.FileDir = path }
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
