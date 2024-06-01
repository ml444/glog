package config

import (
	"os"
	
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/handler"
)

type ReportOption func(cfg *ReportConfig)

func WithReportBaseLogConfig(baseLogConfig *BaseLogConfig) ReportOption {
	return func(cfg *ReportConfig) {
		cfg.BaseLogConfig = baseLogConfig
	}
}

type ReportConfig struct {
	*BaseLogConfig
}

func NewDefaultReportConfig() *ReportConfig {
	curDir, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}
	defaultReportLogDir := curDir
	
	fileConfig := handler.NewFileConfig(
		handler.WithFileRotatorType(handler.FileRotatorTypeSize),
		handler.WithFileDir(defaultReportLogDir),
		handler.WithFileMaxFileSize(defaultMaxFileSize),
		handler.WithFileBulkWriteSize(defaultFileBulkWriteSize),
		handler.WithFileFileSuffix(defaultReportFileSuffix),
		handler.WithFileErrCallback(defaultFileErrCallback),
	)
	
	handlerConfig := handler.NewConfig(
		handler.WithFileConfig(fileConfig),
		handler.WithFormatConfig(formatter.NewConfig()),
	)
	
	return &ReportConfig{
		BaseLogConfig: NewBaseLogConfig(
			WithCacheSize(defaultCacheSize),
			WithHandlerConfig(handlerConfig),
		),
	}
}

func NewReportConfig(opts ...ReportOption) *ReportConfig {
	cfg := NewDefaultReportConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	
	return cfg
}
