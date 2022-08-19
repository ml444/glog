package log

import (
	"fmt"
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/engines"
	"github.com/ml444/glog/message"
	"github.com/ml444/glog/util"
	"github.com/petermattis/goid"
	"runtime"
	"time"

	"os"

	"github.com/ml444/glog/levels"
)

type ILogger interface {
	GetLevel() levels.LogLevel
	SetLevel(levels.LogLevel)

	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Print(...interface{})
	Fatal(...interface{})
	Panic(...interface{})

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Printf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	Panicf(template string, args ...interface{})

	Stop() error
}

var pid = 0
var ip string
var hostName string

func init() {
	pid = os.Getpid()
	hostName, _ = os.Hostname()
	ip, _ = util.GetFirstLocalIp()
}

type Logger struct {
	// module name
	Name   string
	Level  levels.LogLevel
	engine engines.IEngine

	TradeIdFunc FieldFunc

	// Function to exit the application, defaults to `os.Exit()`
	ExitFunc       ExitFunc
	ExitOnFatal    bool
	ExitOnPanic    bool
	IsRecordCaller bool

	isStop bool
}

type FieldFunc func(...interface{}) string
type ExitFunc func(int)

// NewLogger returns a new ILogger
func NewLogger(cfg *config.Config) (*Logger, error) {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	l := Logger{
		Name:           cfg.LoggerName,
		Level:          cfg.LoggerLevel,
		ExitFunc:       os.Exit,
		engine:         engines.NewChanEngine(cfg),
		IsRecordCaller: cfg.IsRecordCaller,
	}
	err := l.init()
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (l *Logger) init() error {
	return l.engine.Init()
}

func (l *Logger) Stop() error {
	l.isStop = true
	return l.engine.Stop()
}

func (l *Logger) send(level levels.LogLevel, msg interface{}) {
	if l.isStop {
		return
	}
	routineId := goid.Get()
	entry := &message.Entry{
		LogName:   l.Name,
		HostName:  hostName,
		Ip:        ip,
		Pid:       pid,
		RoutineId: routineId,
		Message:   msg,
		Time:      time.Now(),
		Level:     level,
		//ErrMsg:    "",
	}
	if l.TradeIdFunc != nil {
		entry.TradeId = l.TradeIdFunc()
	}

	if l.IsRecordCaller {
		entry.Caller = message.GetCaller()
	}
	l.engine.Send(entry)

}

func (l *Logger) IsLevelEnabled(lvl levels.LogLevel) bool {
	return l.GetLevel() < lvl
}
func (l *Logger) GetLevel() levels.LogLevel {
	return l.Level
}
func (l *Logger) SetLevel(lvl levels.LogLevel) {
	l.Level = lvl
}

func (l *Logger) Log(lvl levels.LogLevel, args ...interface{}) {
	if lvl < l.Level {
		return
	}
	msg := fmt.Sprint(args...)
	l.send(lvl, msg)
	l.AfterLog(lvl)
}

func (l *Logger) Logf(lvl levels.LogLevel, template string, args ...interface{}) {
	if lvl < l.Level {
		return
	}

	msg := template
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(template, args...)
	}
	l.send(lvl, msg)
	l.AfterLog(lvl)
}

func (l *Logger) AfterLog(lvl levels.LogLevel) {
	if lvl == levels.FatalLevel || lvl == levels.PanicLevel {
		l.PrintStack(4)
	}
	if l.ExitOnFatal && lvl == levels.FatalLevel {
		err := l.Stop()
		if err != nil {
			println(err)
		}
		l.ExitFunc(-1)
	}
	if l.ExitOnPanic && lvl == levels.PanicLevel {
		err := l.Stop()
		if err != nil {
			println(err)
		}
		l.ExitFunc(-1)
	}
}

func (l *Logger) Debug(args ...interface{}) { l.Log(levels.DebugLevel, args...) }
func (l *Logger) Info(args ...interface{})  { l.Log(levels.InfoLevel, args...) }
func (l *Logger) Warn(args ...interface{})  { l.Log(levels.WarnLevel, args...) }
func (l *Logger) Error(args ...interface{}) { l.Log(levels.ErrorLevel, args...) }
func (l *Logger) Print(args ...interface{}) { l.Log(levels.PrintLevel, args...) }
func (l *Logger) Fatal(args ...interface{}) { l.Log(levels.FatalLevel, args...) }
func (l *Logger) Panic(args ...interface{}) { l.Log(levels.FatalLevel, args...) }

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.Logf(levels.DebugLevel, template, args...)
}
func (l *Logger) Infof(template string, args ...interface{}) {
	l.Logf(levels.InfoLevel, template, args...)
}
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.Logf(levels.WarnLevel, template, args...)
}
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.Logf(levels.ErrorLevel, template, args...)
}
func (l *Logger) Printf(template string, args ...interface{}) {
	l.Logf(levels.PrintLevel, template, args...)
}
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.Logf(levels.PanicLevel, template, args...)
}
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.Logf(levels.FatalLevel, template, args...)
}

func (l *Logger) PrintStack(skip int) {
	for ; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		name := runtime.FuncForPC(pc)
		if name.Name() == "runtime.goexit" {
			break
		}
		l.Errorf("#STACK: %s %s:%d", name.Name(), file, line)
	}
}
