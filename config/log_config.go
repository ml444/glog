package config

import (
	"os"
	"strings"
	
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/level"
)

type LogOption func(cfg *LogConfig)

func WithLogName(name string) LogOption {
	return func(cfg *LogConfig) {
		cfg.Name = name
	}
}

func WithLogBaseLogConfig(baseLogConfig *BaseLogConfig) LogOption {
	return func(cfg *LogConfig) {
		cfg.BaseLogConfig = baseLogConfig
	}
}

type LogConfig struct {
	*BaseLogConfig
	Name string
}

func NewDefaultLogConfig() *LogConfig {
	curDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}
	defaultLogDir := curDir
	l := strings.Split(curDir, string(os.PathSeparator))
	
	fileConfig := handler.NewFileConfig(
		handler.WithFileRotatorType(handler.FileRotatorTypeTimeAndSize),
		handler.WithFileDir(defaultLogDir),
		handler.WithFileMaxFileSize(defaultMaxFileSize),
		handler.WithFileBulkWriteSize(defaultFileBulkWriteSize),
		handler.WithFileTimeSuffixFmt(defaultFileTimeSuffixFmt),
		handler.WithFileReMatch(defaultFileReMatch),
		handler.WithFileErrCallback(defaultFileErrCallback),
	)
	
	formatConfig := formatter.NewConfig(
		formatter.WithFormatterType(formatter.TypeText),
		formatter.WithPatternStyle(PatternTemplateWithDefault),
	)
	
	handlerConfig := handler.NewConfig(
		handler.WithFileConfig(fileConfig),
		handler.WithType(handler.TypeStdout),
		handler.WithFormatConfig(formatConfig),
	)
	
	return &LogConfig{
		Name: l[len(l)-1],
		BaseLogConfig: &BaseLogConfig{
			CacheSize: defaultCacheSize,
			Level:     level.PrintLevel,
			Config:    handlerConfig,
		},
	}
}

func NewLogConfig(opts ...LogOption) *LogConfig {
	cfg := NewDefaultLogConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	
	return cfg
}
