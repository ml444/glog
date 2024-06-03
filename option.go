package log

import "github.com/ml444/glog/message"

type OptionFunc func(config *Config)

// SetLoggerName Set the name of the logger, the default is the name of the program.
func SetLoggerName(name string) OptionFunc {
	return func(cfg *Config) {
		cfg.LoggerName = name
	}
}

func SetLoggerLevel(lvl Level) OptionFunc {
	return func(cfg *Config) { cfg.LoggerLevel = lvl }
}

func SetThrowOnLevel(lvl Level) OptionFunc {
	return func(cfg *Config) { cfg.ThrowOnLevel = lvl }
}

func SetDisableRecordCaller() OptionFunc {
	return func(cfg *Config) { cfg.DisableRecordCaller = true }
}

func SetWorkerConfigs(list ...*WorkerConfig) OptionFunc {
	return func(cfg *Config) { cfg.WorkerConfigList = list }
}

func SetExitFunc(fn func(code int)) OptionFunc {
	return func(cfg *Config) { cfg.ExitFunc = fn }
}

func SetOnError(fn func(v interface{}, err error)) OptionFunc {
	return func(cfg *Config) { cfg.OnError = fn }
}

func SetTraceIDFunc(fn func(entry *message.Entry) string) OptionFunc {
	return func(cfg *Config) { cfg.TraceIDFunc = fn }
}
