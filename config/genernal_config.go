package config

import (
	"os"
	
	"github.com/ml444/glog/message"
)

type GeneralOpt func(cfg *GeneralConfig)

func WithExitOnFatal() GeneralOpt {
	return func(cfg *GeneralConfig) {
		cfg.ExitOnFatal = true
	}
}

func WithThrowOnPanic() GeneralOpt {
	return func(cfg *GeneralConfig) {
		cfg.ThrowOnPanic = true
	}
}

func WithRecordCaller() GeneralOpt {
	return func(cfg *GeneralConfig) {
		cfg.IsRecordCaller = true
	}
}

func WithEnableReport() GeneralOpt {
	return func(cfg *GeneralConfig) {
		cfg.EnableReport = true
	}
}

func WithExitFunc(exitFunc func(code int)) GeneralOpt {
	return func(cfg *GeneralConfig) {
		cfg.ExitFunc = exitFunc
	}
}

func WithTradeIDFunc(tradeIDFunc func(entry *message.Entry) string) GeneralOpt {
	return func(cfg *GeneralConfig) {
		cfg.TradeIDFunc = tradeIDFunc
	}
}

func WithOnError(onError func(msg *message.Entry, err error)) GeneralOpt {
	return func(cfg *GeneralConfig) {
		cfg.OnError = onError
	}
}

type GeneralConfig struct {
	ExitOnFatal    bool
	ThrowOnPanic   bool
	IsRecordCaller bool
	EnableReport   bool
	
	ExitFunc    func(code int)
	TradeIDFunc func(entry *message.Entry) string
	OnError     func(msg *message.Entry, err error)
}

func NewDefaultGeneralConfig() *GeneralConfig {
	return &GeneralConfig{
		ExitFunc: os.Exit,
	}
}

func NewGeneralConfig(opts ...GeneralOpt) *GeneralConfig {
	cfg := NewDefaultGeneralConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	
	return cfg
}
