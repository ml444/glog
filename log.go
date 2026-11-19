package log

import (
	"os"
	"sync"
	"sync/atomic"

	"github.com/ml444/glog/level"
)

type Level = level.LogLevel

const (
	NoneLevel Level = iota
	DebugLevel
	PrintLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

var (
	loggerMu    sync.Mutex
	loggerState atomic.Value
	Conf        = NewDefaultConfig()
)

type loggerHolder struct {
	activeState int64
	logger      ILogger
	drained     chan struct{}
}

const loggerRetired int64 = -1 << 63

func newLoggerHolder(logger ILogger) *loggerHolder {
	return &loggerHolder{logger: logger, drained: make(chan struct{})}
}

func acquireLogger() *loggerHolder {
	for {
		holder := loggerState.Load().(*loggerHolder)
		state := atomic.LoadInt64(&holder.activeState)
		if state >= 0 && atomic.CompareAndSwapInt64(&holder.activeState, state, state+1) {
			return holder
		}
	}
}

func (holder *loggerHolder) release() {
	if atomic.AddInt64(&holder.activeState, -1) == loggerRetired {
		close(holder.drained)
	}
}

func (holder *loggerHolder) retire() error {
	if atomic.AddInt64(&holder.activeState, loggerRetired) == loggerRetired {
		close(holder.drained)
	}
	<-holder.drained
	if holder.logger != nil {
		return holder.logger.Stop()
	}
	return nil
}

func init() {
	logger, err := NewLogger(Conf)
	if err != nil {
		panic(err.Error())
	}
	loggerState.Store(newLoggerHolder(logger))
	registerShutdownOnSignals()
}

func InitLog(opts ...OptionFunc) error {
	old, onError, err := installLogger(opts...)
	if err != nil {
		return err
	}
	if err := old.retire(); err != nil && onError != nil {
		onError(nil, err)
	}
	return nil
}

func installLogger(opts ...OptionFunc) (*loggerHolder, func(interface{}, error), error) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	cfg := Conf.clone()
	for _, optionFunc := range opts {
		if optionFunc != nil {
			optionFunc(cfg)
		}
	}
	l, err := NewLogger(cfg)
	if err != nil {
		return nil, nil, err
	}
	old := loggerState.Load().(*loggerHolder)
	loggerState.Store(newLoggerHolder(l))
	Conf = cfg
	return old, cfg.OnError, nil
}

func SetLogger(g ILogger) {
	loggerMu.Lock()
	loggerState.Store(newLoggerHolder(g))
	loggerMu.Unlock()
}

func GetLogger() ILogger {
	return loggerState.Load().(*loggerHolder).logger
}

func Stats() LoggerStats {
	holder := acquireLogger()
	defer holder.release()
	if l, ok := holder.logger.(interface{ Stats() LoggerStats }); ok {
		return l.Stats()
	}
	return LoggerStats{}
}

func Debug(args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Debug(args...)
}

func Info(args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Info(args...)
}

func Warn(args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Warn(args...)
}

func Error(args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Error(args...)
}

func Print(args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Print(args...)
}

func Panic(args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Panic(args...)
}

func Fatal(args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Fatal(args...)
}

func Debugf(template string, args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Errorf(template, args...)
}

func Printf(template string, args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Printf(template, args...)
}

func Panicf(template string, args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Panicf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	holder.logger.Fatalf(template, args...)
}

// Enabled reports whether the package logger accepts lvl. Use it to avoid
// building expensive log arguments when that level is disabled.
func Enabled(lvl Level) bool {
	holder := acquireLogger()
	defer holder.release()
	l := holder.logger
	if enabledLogger, ok := l.(interface{ Enabled(Level) bool }); ok {
		return enabledLogger.Enabled(lvl)
	}
	return l != nil && lvl >= l.GetLevel()
}

func Log(lvl Level, args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	l := holder.logger
	if levelLogger, ok := l.(interface {
		Log(Level, ...interface{})
	}); ok {
		levelLogger.Log(lvl, args...)
		return
	}
	logWithInterface(l, lvl, args...)
}

func Logf(lvl Level, template string, args ...interface{}) {
	holder := acquireLogger()
	defer holder.release()
	l := holder.logger
	if levelLogger, ok := l.(interface {
		Logf(Level, string, ...interface{})
	}); ok {
		levelLogger.Logf(lvl, template, args...)
		return
	}
	logfWithInterface(l, lvl, template, args...)
}

func logWithInterface(l ILogger, lvl Level, args ...interface{}) {
	switch lvl {
	case DebugLevel:
		l.Debug(args...)
	case PrintLevel:
		l.Print(args...)
	case InfoLevel:
		l.Info(args...)
	case WarnLevel:
		l.Warn(args...)
	case ErrorLevel:
		l.Error(args...)
	case PanicLevel:
		l.Panic(args...)
	case FatalLevel:
		l.Fatal(args...)
	}
}

func logfWithInterface(l ILogger, lvl Level, template string, args ...interface{}) {
	switch lvl {
	case DebugLevel:
		l.Debugf(template, args...)
	case PrintLevel:
		l.Printf(template, args...)
	case InfoLevel:
		l.Infof(template, args...)
	case WarnLevel:
		l.Warnf(template, args...)
	case ErrorLevel:
		l.Errorf(template, args...)
	case PanicLevel:
		l.Panicf(template, args...)
	case FatalLevel:
		l.Fatalf(template, args...)
	}
}

func Stop() {
	holder := acquireLogger()
	defer holder.release()
	if l := holder.logger; l != nil {
		if err := l.Stop(); err != nil {
			println(err.Error())
			return
		}
	}
}

// Exit calls Stop to flush asynchronous loggers, then terminates the process with the given code.
// Use it instead of os.Exit when you need a guaranteed flush before exit; registerShutdownOnSignals
// does not intercept os.Exit.
func Exit(code int) {
	Stop()
	os.Exit(code)
}
