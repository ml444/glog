package handler

import (
	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
)

type Opt func(cfg *Config)

func WithExternalHandler(handler IHandler) Opt {
	return func(cfg *Config) {
		cfg.ExternalHandler = handler
	}
}

func WithType(typ Type) Opt {
	return func(cfg *Config) {
		cfg.Type = typ
	}
}

func WithFileConfig(fileConfig *FileConfig) Opt {
	return func(cfg *Config) {
		cfg.Type = TypeFile
		cfg.File = fileConfig
	}
}

func WithStreamConfig(streamConfig *StreamConfig) Opt {
	return func(cfg *Config) {
		cfg.Type = TypeStream
		cfg.Stream = streamConfig
	}
}

func WithSyslogConfig(syslogConfig *SyslogConfig) Opt {
	return func(cfg *Config) {
		cfg.Type = TypeSyslog
		cfg.Syslog = syslogConfig
	}
}

func WithFormatConfig(formatConfig *formatter.Config) Opt {
	return func(cfg *Config) {
		cfg.FormatConfig = formatConfig
	}
}

func WithFilter(filter filter.IFilter) Opt {
	return func(cfg *Config) {
		cfg.Filter = filter
	}
}
