package log

import (
	"fmt"
	"github.com/ml444/glog/util"
	"runtime"
	"time"

	"github.com/ml444/glog/config"
	"github.com/ml444/glog/engines"
	"github.com/ml444/glog/message"
	"github.com/petermattis/goid"

	"github.com/ml444/glog/levels"
)

type StdLogger interface {
	Print(...interface{})
	Println(...interface{})
	Printf(string, ...interface{})

	Fatal(...interface{})
	Fatalln(...interface{})
	Fatalf(string, ...interface{})

	Panic(...interface{})
	Panicln(...interface{})
	Panicf(string, ...interface{})
}

type ILogger interface {
	GetLevel() levels.LogLevel
	SetLevel(levels.LogLevel)
	EnabledLevel(level levels.LogLevel) bool

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

type Logger struct {
	Name   string
	Level  levels.LogLevel
	engine engines.IEngine

	TradeIDFunc    func(entry *message.Entry) string
	ExitFunc       func(code int) // Function to exit the application, defaults to `os.Exit()`
	ExitOnFatal    bool
	IsRecordCaller bool
	isStop         bool
}

type FieldFunc func(entry *message.Entry) string

var _ StdLogger = &Logger{}
var _ ILogger = &Logger{}

// NewLogger returns a new ILogger
func NewLogger(cfg *config.Config) (*Logger, error) {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	l := Logger{
		Name:           cfg.LoggerName,
		Level:          cfg.LoggerLevel,
		ExitFunc:       cfg.ExitFunc,
		engine:         engines.NewEngine(cfg.EngineType),
		IsRecordCaller: cfg.IsRecordCaller,
		TradeIDFunc:    cfg.TradeIDFunc,
	}
	err := l.init()
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (l *Logger) init() (err error) {
	err = l.engine.Init()
	if err != nil {
		return err
	}
	err = l.engine.Start()
	if err != nil {
		return err
	}
	return nil
}
func (l *Logger) send(level levels.LogLevel, msg interface{}) {
	if l.isStop {
		return
	}
	routineID := goid.Get()
	entry := &message.Entry{
		LogName:   l.Name,
		RoutineId: routineID,
		Message:   msg,
		Time:      time.Now(),
		Level:     level,
		//ErrMsg:    "",
	}
	if l.TradeIDFunc != nil {
		entry.TradeId = l.TradeIDFunc(entry)
	}

	if l.IsRecordCaller {
		entry.Caller = util.GetCaller()
	}
	l.engine.Send(entry)
}
func (l *Logger) log(lvl levels.LogLevel, args ...interface{}) {
	if lvl < l.Level {
		return
	}
	msg := fmt.Sprint(args...)
	l.send(lvl, msg)
	l.after(lvl)
}
func (l *Logger) logf(lvl levels.LogLevel, template string, args ...interface{}) {
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
	l.after(lvl)
}
func (l *Logger) after(lvl levels.LogLevel) {
	if lvl == levels.FatalLevel || lvl == levels.PanicLevel {
		l.printStack(2, lvl)
	}
	if (lvl == levels.PanicLevel) || (l.ExitOnFatal && lvl == levels.FatalLevel) {
		err := l.Stop()
		if err != nil {
			println(err)
		}
		l.ExitFunc(-1)
	}
}
func (l *Logger) printStack(callDepth int, lvl levels.LogLevel) {
	for ; ; callDepth++ {
		pc, file, line, ok := runtime.Caller(callDepth)
		if !ok {
			break
		}
		name := runtime.FuncForPC(pc)
		if name.Name() == "runtime.goexit" {
			break
		}
		l.send(lvl, fmt.Sprintf("#STACK: %s %s:%d", name.Name(), file, line))
	}
}

func (l *Logger) EnabledLevel(lvl levels.LogLevel) bool {
	return l.GetLevel() < lvl
}
func (l *Logger) GetLevel() levels.LogLevel {
	return l.Level
}
func (l *Logger) SetLevel(lvl levels.LogLevel) {
	l.Level = lvl
}

func (l *Logger) Debug(args ...interface{}) { l.log(levels.DebugLevel, args...) }
func (l *Logger) Info(args ...interface{})  { l.log(levels.InfoLevel, args...) }
func (l *Logger) Warn(args ...interface{})  { l.log(levels.WarnLevel, args...) }
func (l *Logger) Error(args ...interface{}) { l.log(levels.ErrorLevel, args...) }

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.logf(levels.DebugLevel, template, args...)
}
func (l *Logger) Infof(template string, args ...interface{}) {
	l.logf(levels.InfoLevel, template, args...)
}
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.logf(levels.WarnLevel, template, args...)
}
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.logf(levels.ErrorLevel, template, args...)
}

func (l *Logger) Print(args ...interface{}) { l.log(levels.PrintLevel, args...) }
func (l *Logger) Fatal(args ...interface{}) { l.log(levels.FatalLevel, args...) }
func (l *Logger) Panic(args ...interface{}) { l.log(levels.PanicLevel, args...) }

func (l *Logger) Println(args ...interface{}) { l.log(levels.PrintLevel, args...) }
func (l *Logger) Fatalln(args ...interface{}) { l.log(levels.FatalLevel, args...) }
func (l *Logger) Panicln(args ...interface{}) { l.log(levels.PanicLevel, args...) }

func (l *Logger) Printf(template string, args ...interface{}) {
	l.logf(levels.PrintLevel, template, args...)
}
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.logf(levels.PanicLevel, template, args...)
}
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.logf(levels.FatalLevel, template, args...)
}

func (l *Logger) Stop() error {
	l.isStop = true
	return l.engine.Stop()
}
