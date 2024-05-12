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

func SetHandler2Logger(h handler.IHandler) OptionFunc {
	return func(cfg *Config) { cfg.LogHandlerConfig.ExternalHandler = h }
}

func SetHandler2Report(h handler.IHandler) OptionFunc {
	return func(cfg *Config) { cfg.ReportHandlerConfig.ExternalHandler = h }
}

func SetHandlerCfg2Logger(hcfg handler.HandlerConfig) OptionFunc {
	return func(cfg *Config) { cfg.LogHandlerConfig = hcfg }
}

func SetHandlerCfg2Report(hcfg handler.HandlerConfig) OptionFunc {
	return func(cfg *Config) { cfg.ReportHandlerConfig = hcfg }
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

func SetSyslogHandlerCfg2Logger(syslogCfg handler.SyslogHandlerConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.Syslog = syslogCfg
	}
}

func SetSyslogHandlerCfg2Report(syslogCfg handler.SyslogHandlerConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.ReportHandlerConfig.Syslog = syslogCfg
	}
}

func SetFormatterCfg2Logger(formatterCfg formatter.FormatterConfig) OptionFunc {
	return func(cfg *Config) {
		cfg.LogHandlerConfig.Formatter = formatterCfg
	}
}

func SetFormatterCfg2Report(formatterCfg formatter.FormatterConfig) OptionFunc {
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
