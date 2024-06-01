package handler

import (
	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
)

type Option func(cfg *Config)

func WithExternalHandler(handler IHandler) Option {
	return func(cfg *Config) {
		cfg.ExternalHandler = handler
	}
}

func WithType(typ Type) Option {
	return func(cfg *Config) {
		cfg.Type = typ
	}
}

func WithFileConfig(fileConfig *FileConfig) Option {
	return func(cfg *Config) {
		cfg.Type = TypeFile
		cfg.File = fileConfig
	}
}

func WithStreamConfig(streamConfig *StreamConfig) Option {
	return func(cfg *Config) {
		cfg.Type = TypeStream
		cfg.Stream = streamConfig
	}
}

func WithSyslogConfig(syslogConfig *SyslogConfig) Option {
	return func(cfg *Config) {
		cfg.Type = TypeSyslog
		cfg.Syslog = syslogConfig
	}
}

func WithFormatConfig(formatConfig *formatter.Config) Option {
	return func(cfg *Config) {
		cfg.FormatConfig = formatConfig
	}
}

func WithFilter(filter filter.IFilter) Option {
	return func(cfg *Config) {
		cfg.Filter = filter
	}
}
