package log

import (
	"errors"
	"github.com/ml444/glog/config"
)

var logger ILogger
var Config *config.Config

func init() {
	if logger != nil {
		return
	}
	if Config == nil {
		Config = config.NewDefaultConfig()
	}
	logger, _ = NewLogger(Config)
}

func InitLog(opts ...config.OptionFunc) error {
	if Config == nil {
		Config = config.NewDefaultConfig()
	}
	for _, optionFunc := range opts {
		err := optionFunc(Config)
		if err != nil {
			return err
		}
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

func Debug(args ...interface{}) { logger.Debug(args...) }
func Info(args ...interface{})  { logger.Info(args...) }
func Error(args ...interface{}) { logger.Error(args...) }
func Warn(args ...interface{})  { logger.Warn(args...) }
func Print(args ...interface{}) { logger.Print(args...) }
func Panic(args ...interface{}) { logger.Panic(args...) }
func Fatal(args ...interface{}) { logger.Fatal(args...) }

func Debugf(template string, args ...interface{}) { logger.Debugf(template, args...) }
func Infof(template string, args ...interface{})  { logger.Infof(template, args...) }
func Errorf(template string, args ...interface{}) { logger.Errorf(template, args...) }
func Warnf(template string, args ...interface{})  { logger.Warnf(template, args...) }
func Printf(template string, args ...interface{}) { logger.Printf(template, args...) }
func Panicf(template string, args ...interface{}) { logger.Panicf(template, args...) }
func Fatalf(template string, args ...interface{}) { logger.Fatalf(template, args...) }

func Exit() error {
	if logger != nil {
		return logger.Stop()
	}
	return errors.New("logger not open")
}
