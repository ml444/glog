package config

import (
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/level"
)

type BaseLogOption func(cfg *BaseLogConfig)

func WithCacheSize(size int) BaseLogOption {
	return func(cfg *BaseLogConfig) {
		cfg.CacheSize = size
	}
}

func WithLevel(lvl level.LogLevel) BaseLogOption {
	return func(cfg *BaseLogConfig) {
		cfg.Level = lvl
	}
}

func WithHandlerConfig(handlerConfig *handler.Config) BaseLogOption {
	return func(cfg *BaseLogConfig) {
		cfg.Config = handlerConfig
	}
}

type BaseLogConfig struct {
	CacheSize int
	Level     level.LogLevel
	Config    *handler.Config
}

func NewDefaultBaseLogConfig() *BaseLogConfig {
	return &BaseLogConfig{
		Level:  level.InfoLevel,
		Config: handler.NewConfig(),
	}
}

func NewBaseLogConfig(opts ...BaseLogOption) *BaseLogConfig {
	cfg := NewDefaultBaseLogConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	
	return cfg
}
