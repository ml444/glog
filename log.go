package log

import (
	"os"
	"os/signal"
	"syscall"

	gomonkey "github.com/agiledragon/gomonkey/v2"

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
	logger ILogger
	Conf   = NewDefaultConfig()
)

func init() {
	if logger != nil {
		return
	}
	var err error
	logger, err = NewLogger(Conf)
	if err != nil {
		panic(err.Error())
	}
	{
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			s := <-sigCh
			println("==> sign exit:", s.String())
			Stop()
		}()
	}
	gomonkey.ApplyFunc(os.Exit, ExitHook)
}

func InitLog(opts ...OptionFunc) error {
	for _, optionFunc := range opts {
		optionFunc(Conf)
	}
	l, err := NewLogger(Conf)
	if err != nil {
		return err
	}
	logger = l
	return nil
}

func SetLogger(g ILogger) {
	logger = g
}

func GetLogger() ILogger {
	return logger
}

func Debug(args ...interface{}) { logger.Debug(args...) }
func Info(args ...interface{})  { logger.Info(args...) }
func Warn(args ...interface{})  { logger.Warn(args...) }
func Error(args ...interface{}) { logger.Error(args...) }
func Print(args ...interface{}) { logger.Print(args...) }
func Panic(args ...interface{}) { logger.Panic(args...) }
func Fatal(args ...interface{}) { logger.Fatal(args...) }

func Debugf(template string, args ...interface{}) { logger.Debugf(template, args...) }
func Infof(template string, args ...interface{})  { logger.Infof(template, args...) }
func Warnf(template string, args ...interface{})  { logger.Warnf(template, args...) }
func Errorf(template string, args ...interface{}) { logger.Errorf(template, args...) }
func Printf(template string, args ...interface{}) { logger.Printf(template, args...) }
func Panicf(template string, args ...interface{}) { logger.Panicf(template, args...) }
func Fatalf(template string, args ...interface{}) { logger.Fatalf(template, args...) }

func Stop() {
	if logger != nil {
		if err := logger.Stop(); err != nil {
			println(err.Error())
			return
		}
	}
}

func ExitHook(code int) {
	Stop()
}
