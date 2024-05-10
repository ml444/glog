package log

import (
	"os"
	"os/signal"
	"syscall"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/ml444/glog/config"
)

var (
	logger ILogger
	Config = config.NewDefaultConfig()
)

func init() {
	if logger != nil {
		return
	}
	var err error
	logger, err = NewLogger(Config)
	if err != nil {
		println(err)
		return
	}
	{
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			s := <-sigCh
			println("==> sign exit:", s.String())
			Exit()
		}()
	}
	gomonkey.ApplyFunc(os.Exit, ExitHook)
}

func InitLog(opts ...config.OptionFunc) error {
	for _, optionFunc := range opts {
		optionFunc(Config)
	}
	l, err := NewLogger(Config)
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

func Exit() {
	if logger != nil {
		if err := logger.Stop(); err != nil {
			println(err)
			return
		}
	}
}

func ExitHook(code int) {
	Exit()
}