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
	//SetField(key string, fn FieldFunc)

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

	Init() error
	Stop()
	Sync() error
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
	IsRecordCaller bool
}

type FieldFunc func(...interface{}) string
type ExitFunc func(int)

func GetLogger(name string) ILogger {
	cfg := config.NewDefaultConfig()
	cfg.Logger.Name = name
	return NewLogger(cfg)
}

// NewLogger returns a new ILogger
func NewLogger(cfg *config.Config) *Logger {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	lCfg := cfg.Logger
	return &Logger{
		Name:        lCfg.Name,
		Level:       lCfg.Level,
		ExitOnFatal: true,
		ExitFunc:    os.Exit,
		engine:      engines.NewChanEngine(cfg),
		IsRecordCaller: lCfg.IsRecordCaller,
	}
}

func (l *Logger) Init() error {
	return l.engine.Init()
}

func (l *Logger) Stop() {
	l.engine.Stop()
}

func (l Logger) Sync() error {
	return l.engine.Sync()
}

func (l *Logger) write(level levels.LogLevel, msg interface{}) {
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
		ErrMsg:    "",
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
	l.write(lvl, msg)

	if lvl == levels.FatalLevel || lvl == levels.PanicLevel || lvl == levels.DPanicLevel {
		l.PrintStack(4)
	}
	if l.ExitOnFatal && lvl == levels.FatalLevel {
		l.ExitFunc(1)
		//os.Exit(1)
	}
	if lvl == levels.PanicLevel {
		panic(msg)
	}
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
	l.write(lvl, msg)

	if lvl == levels.FatalLevel || lvl == levels.PanicLevel || lvl == levels.DPanicLevel {
		l.PrintStack(4)
	}
	if l.ExitOnFatal && lvl == levels.FatalLevel {
		l.ExitFunc(1)
		//os.Exit(1)
	}
	if lvl == levels.PanicLevel {
		panic("")
	}
}

func (l *Logger) Debug(args ...interface{}) { l.Log(levels.DebugLevel, args...) }
func (l *Logger) Info(args ...interface{})  { l.Log(levels.InfoLevel, args...) }
func (l *Logger) Warn(args ...interface{})  { l.Log(levels.WarnLevel, args...) }
func (l *Logger) Error(args ...interface{}) { l.Log(levels.ErrorLevel, args...) }
func (l *Logger) Print(args ...interface{}) { l.Log(levels.InfoLevel, args...) }
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
	l.Logf(levels.InfoLevel, template, args...)
}
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.Logf(levels.DPanicLevel, template, args...)
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
