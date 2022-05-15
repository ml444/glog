package config

import "github.com/ml444/glog/levels"

type OptionFunc func(config *Config) error

func SetLoggerName(name string) OptionFunc {
	return func(cfg *Config) error {
		cfg.LoggerName = name
		return nil
	}
}

func SetLoggerLevel(level levels.LogLevel) OptionFunc {
	return func(cfg *Config) error {
		cfg.LoggerLevel = level
		return nil
	}
}

func SetReportLevel(level levels.LogLevel) OptionFunc {
	return func(cfg *Config) error {
		cfg.ReportLevel = level
		return nil
	}
}

func SetEnableReport(enable bool) OptionFunc {
	return func(cfg *Config) error {
		cfg.EnableReport = enable
		return nil
	}
}

func SetCacheSize(size int, isSetReport bool) OptionFunc {
	return func(cfg *Config) error {
		if isSetReport {
			cfg.ReportCacheSize = size
		} else {
			cfg.LoggerCacheSize = size
		}
		return nil
	}
}


func SetFileName(name string, isSetReport bool) OptionFunc {
	return func(cfg *Config) error {
		if isSetReport {
			cfg.Handler.ReportHandlerConfig.File.FileName = name
		} else {
			cfg.Handler.LogHandlerConfig.File.FileName = name
		}
		return nil
	}
}

func SetFileDir(path string, isSetReport bool) OptionFunc {
	return func(cfg *Config) error {
		if isSetReport {
			cfg.Handler.ReportHandlerConfig.File.FileDir = path
		} else {
			cfg.Handler.LogHandlerConfig.File.FileDir = path
		}
		return nil
	}
}

func SetFileBackupCount(count int, isSetReport bool) OptionFunc {
	return func(cfg *Config) error {
		if isSetReport {
			cfg.Handler.ReportHandlerConfig.File.BackupCount = count
		} else {
			cfg.Handler.LogHandlerConfig.File.BackupCount = count
		}
		return nil
	}
}
func SetFileMaxSize(size int64, isSetReport bool) OptionFunc {
	return func(cfg *Config) error {
		if isSetReport {
			cfg.Handler.ReportHandlerConfig.File.MaxFileSize = size
		} else {
			cfg.Handler.LogHandlerConfig.File.MaxFileSize = size
		}
		return nil
	}
}
func SetFileWhen(when uint8, isSetReport bool) OptionFunc {
	return func(cfg *Config) error {
		if isSetReport {
			cfg.Handler.ReportHandlerConfig.File.When = when
		} else {
			cfg.Handler.LogHandlerConfig.File.When = when
		}
		return nil
	}
}

func SetFileRematch(pattern string, isSetReport bool) OptionFunc {
	return func(cfg *Config) error {
		if isSetReport {
			cfg.Handler.ReportHandlerConfig.File.ReMatch = pattern
		} else {
			cfg.Handler.LogHandlerConfig.File.ReMatch = pattern
		}
		return nil
	}
}
func SetFileTimeFmtSuffix(timeFmt string, isSetReport bool) OptionFunc {
	return func(cfg *Config) error {
		if isSetReport {
			cfg.Handler.ReportHandlerConfig.File.SuffixFmt = timeFmt
		} else {
			cfg.Handler.LogHandlerConfig.File.SuffixFmt = timeFmt
		}
		return nil
	}
}
