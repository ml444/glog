package log

import (
	"fmt"
	"runtime"
	"strings"
	"time"
	
	"github.com/ml444/glog/util"
	
	"github.com/petermattis/goid"
	
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/engine"
	"github.com/ml444/glog/message"
	
	"github.com/ml444/glog/level"
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
	GetLoggerName() string
	SetLoggerName(string)
	
	GetLevel() level.LogLevel
	SetLevel(level.LogLevel)
	EnabledLevel(lvl level.LogLevel) bool
	
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
	Level  level.LogLevel
	engine engine.IEngine
	
	TradeIDFunc    func(entry *message.Entry) string
	ExitFunc       func(code int) // Function to exit the application, defaults to `os.Exit()`
	ExitOnFatal    bool
	ThrowOnPanic   bool
	IsRecordCaller bool
	isStop         bool
}

// type FieldFunc func(entry *message.Entry) string
var (
	_ StdLogger = &Logger{}
	_ ILogger   = &Logger{}
)

// NewLogger returns a new ILogger
func NewLogger(cfg *config.Config) (*Logger, error) {
	if cfg == nil {
		cfg = config.NewConfig()
	}
	l := Logger{
		Name:           cfg.LogConfig.Name,
		Level:          cfg.LogConfig.Level,
		ExitFunc:       cfg.ExitFunc,
		ExitOnFatal:    cfg.ExitOnFatal,
		ThrowOnPanic:   cfg.ThrowOnPanic,
		engine:         engine.NewChannelEngine(cfg),
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
	err = l.engine.Start()
	if err != nil {
		_ = l.Stop()
		return err
	}
	return nil
}

func (l *Logger) send(lvl level.LogLevel, msg string) {
	if l.isStop {
		println("it is stoped, can't send: ", msg)
		return
	}
	routineID := goid.Get()
	entry := &message.Entry{
		LogName:   l.Name,
		RoutineID: routineID,
		Message:   msg,
		Time:      time.Now(),
		Level:     lvl,
	}
	if l.TradeIDFunc != nil {
		entry.TraceID = l.TradeIDFunc(entry)
	}
	
	if l.IsRecordCaller {
		entry.Caller = util.GetCaller()
	}
	l.engine.Send(entry)
}

func (l *Logger) log(lvl level.LogLevel, args ...interface{}) {
	if lvl < l.Level {
		return
	}
	msg := fmt.Sprint(args...)
	l.send(lvl, msg)
	l.after(lvl)
}

func (l *Logger) logf(lvl level.LogLevel, template string, args ...interface{}) {
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

func (l *Logger) after(lvl level.LogLevel) {
	if lvl == level.FatalLevel || lvl == level.PanicLevel {
		l.printStack(4, lvl)
	}
	if (l.ThrowOnPanic && lvl == level.PanicLevel) || (l.ExitOnFatal && lvl == level.FatalLevel) {
		err := l.Stop()
		if err != nil {
			println(err)
		}
		l.ExitFunc(-1)
	}
}

func (l *Logger) printStack(callDepth int, lvl level.LogLevel) {
	buf := new(strings.Builder)
	buf.WriteString("\n")
	for ; ; callDepth++ {
		pc, file, line, ok := runtime.Caller(callDepth)
		if !ok {
			break
		}
		name := runtime.FuncForPC(pc)
		if name.Name() == "runtime.goexit" {
			break
		}
		fmt.Fprintf(buf, "	[STACK]: %s %s:%d\n", name.Name(), file, line)
	}
	l.send(lvl, buf.String())
}

func (l *Logger) GetLoggerName() string {
	return l.Name
}

func (l *Logger) SetLoggerName(name string) {
	l.Name = name
}

func (l *Logger) EnabledLevel(lvl level.LogLevel) bool {
	return l.GetLevel() < lvl
}

func (l *Logger) GetLevel() level.LogLevel {
	return l.Level
}

func (l *Logger) SetLevel(lvl level.LogLevel) {
	l.Level = lvl
}

func (l *Logger) Debug(args ...interface{}) { l.log(level.DebugLevel, args...) }
func (l *Logger) Info(args ...interface{})  { l.log(level.InfoLevel, args...) }
func (l *Logger) Warn(args ...interface{})  { l.log(level.WarnLevel, args...) }
func (l *Logger) Error(args ...interface{}) { l.log(level.ErrorLevel, args...) }

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.logf(level.DebugLevel, template, args...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.logf(level.InfoLevel, template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.logf(level.WarnLevel, template, args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.logf(level.ErrorLevel, template, args...)
}

func (l *Logger) Print(args ...interface{}) { l.log(level.PrintLevel, args...) }
func (l *Logger) Fatal(args ...interface{}) { l.log(level.FatalLevel, args...) }
func (l *Logger) Panic(args ...interface{}) { l.log(level.PanicLevel, args...) }

func (l *Logger) Println(args ...interface{}) { l.log(level.PrintLevel, args...) }
func (l *Logger) Fatalln(args ...interface{}) { l.log(level.FatalLevel, args...) }
func (l *Logger) Panicln(args ...interface{}) { l.log(level.PanicLevel, args...) }

func (l *Logger) Printf(template string, args ...interface{}) {
	l.logf(level.PrintLevel, template, args...)
}

func (l *Logger) Panicf(template string, args ...interface{}) {
	l.logf(level.PanicLevel, template, args...)
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.logf(level.FatalLevel, template, args...)
}

func (l *Logger) Stop() error {
	defer func() {
		l.isStop = true
	}()
	return l.engine.Stop()
}
