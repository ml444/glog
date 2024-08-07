package log

import "github.com/ml444/glog/message"

type OptionFunc func(config *Config)

// SetLoggerName Set the name of the logger, the default is the name of the program.
func SetLoggerName(name string) OptionFunc { return func(cfg *Config) { cfg.LoggerName = name } }

// SetLoggerLevel Set the global log level. Logs below this level will not be processed.
func SetLoggerLevel(lvl Level) OptionFunc {
	return func(cfg *Config) { cfg.LoggerLevel = lvl }
}

// SetThrowOnLevel what level of logging is set here will trigger an exception to be thrown.
func SetThrowOnLevel(lvl Level) OptionFunc {
	return func(cfg *Config) { cfg.ThrowOnLevel = lvl }
}

// SetRecordCaller Enable recording of caller information
func SetRecordCaller(skip int) OptionFunc {
	return func(cfg *Config) { 
		cfg.EnableRecordCaller = true 
		cfg.CallerSkipCount = skip
	}
}

// SetColorRender enable color rendering. Only enabled by default in the text formatter.
func SetColorRender(enable bool) OptionFunc {
	return func(cfg *Config) { cfg.EnableColorRender = &enable }
}

// SetTimeLayout time layout string, for example: "2006-01-02 15:04:05.000"
func SetTimeLayout(layout string) OptionFunc { return func(cfg *Config) { cfg.TimeLayout = layout } }

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
