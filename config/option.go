package config

import "github.com/ml444/glog/levels"

type OptionFunc func(config *Config) error

func SetLoggerName(name string) OptionFunc {
	return func(cfg *Config) error {
		cfg.Logger.Name = name
		return nil
	}
}



func SetLogLevel(level levels.LogLevel) OptionFunc {
	return func(cfg *Config) error {
		cfg.Logger.Level = level
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
