package log

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/util"

	"github.com/petermattis/goid"

	"github.com/ml444/glog/message"
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

	GetLevel() Level
	SetLevel(Level)

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
	Name               string
	Level              Level
	ThrowOnLevel       Level
	ExitFunc           func(code int) // Function to exit the application, defaults to `os.Exit()`
	TraceIDFunc        func(record *message.Record) string
	engine             IEngine
	enableRecordCaller bool
	isStop             bool
}

// type FieldFunc func(record *message.Record) string
var (
	_ StdLogger = &Logger{}
	_ ILogger   = &Logger{}
)

// NewLogger returns a new ILogger
func NewLogger(cfg *Config) (*Logger, error) {
	if cfg == nil {
		cfg = NewDefaultConfig()
	}
	cfg.Check()
	eng, err := NewChannelEngine(cfg)
	if err != nil {
		return nil, err
	}
	l := Logger{
		Name:               cfg.LoggerName,
		Level:              cfg.LoggerLevel,
		ThrowOnLevel:       cfg.ThrowOnLevel,
		ExitFunc:           cfg.ExitFunc,
		TraceIDFunc:        cfg.TraceIDFunc,
		engine:             eng,
		enableRecordCaller: cfg.EnableRecordCaller,
	}
	err = l.init()
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
	formatter.SetLoggerName(l.Name)
	return nil
}

func (l *Logger) newRecord() *message.Record {
	record := message.Record{
		LogName:      l.Name,
		Level:        l.Level,
		RoutineID:    goid.Get(),
		Time:         time.Now(),
		SendCallback: l.send,
	}
	if l.TraceIDFunc != nil {
		record.TraceID = l.TraceIDFunc(&record)
	}

	if l.enableRecordCaller {
		record.Caller = util.GetCaller()
	}
	return &record
}

func (l *Logger) send(record *message.Record) {
	if l.isStop {
		println("it is stoped, can't send: ", record.Message)
		return
	}
	l.engine.Send(record)
}

func (l *Logger) log(lvl Level, args ...interface{}) {
	if lvl < l.Level {
		return
	}
	record := l.newRecord()
	record.Message = fmt.Sprint(args...)
	l.send(record)
	l.after(lvl)
}

func (l *Logger) logf(lvl Level, template string, args ...interface{}) {
	if lvl < l.Level {
		return
	}

	msg := template
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(template, args...)
	}
	record := l.newRecord()
	record.Message = msg
	l.send(record)
	l.after(lvl)
}

func (l *Logger) after(lvl Level) {
	if lvl < PanicLevel {
		return
	}
	l.printStack(4, lvl)
	if l.ThrowOnLevel != NoneLevel && lvl >= l.ThrowOnLevel {
		err := l.Stop()
		if err != nil {
			println(err.Error())
		}
		l.ExitFunc(-1)
	}
}

func (l *Logger) printStack(callDepth int, lvl Level) {
	buf := new(strings.Builder)
	buf.WriteString("\n")
	for ; ; callDepth++ {
		pc, file, line, ok := runtime.Caller(callDepth)
		if !ok {
			break
		}
		funcName := runtime.FuncForPC(pc).Name()
		if funcName == "runtime.goexit" {
			break
		}
		buf.WriteString(fmt.Sprintf("	[STACK]: %s %s:%d\n", funcName, file, line))
	}
	record := l.newRecord()
	record.Message = buf.String()
	l.send(record)
}

func (l *Logger) GetLoggerName() string {
	return l.Name
}

func (l *Logger) SetLoggerName(name string) {
	l.Name = name
}

func (l *Logger) GetLevel() Level {
	return l.Level
}

func (l *Logger) SetLevel(lvl Level) {
	l.Level = lvl
}

func (l *Logger) Debug(args ...interface{}) { l.log(DebugLevel, args...) }
func (l *Logger) Info(args ...interface{})  { l.log(InfoLevel, args...) }
func (l *Logger) Warn(args ...interface{})  { l.log(WarnLevel, args...) }
func (l *Logger) Error(args ...interface{}) { l.log(ErrorLevel, args...) }

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.logf(DebugLevel, template, args...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.logf(InfoLevel, template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.logf(WarnLevel, template, args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.logf(ErrorLevel, template, args...)
}

func (l *Logger) Print(args ...interface{}) { l.log(PrintLevel, args...) }
func (l *Logger) Fatal(args ...interface{}) { l.log(FatalLevel, args...) }
func (l *Logger) Panic(args ...interface{}) { l.log(PanicLevel, args...) }

func (l *Logger) Println(args ...interface{}) { l.log(PrintLevel, args...) }
func (l *Logger) Fatalln(args ...interface{}) { l.log(FatalLevel, args...) }
func (l *Logger) Panicln(args ...interface{}) { l.log(PanicLevel, args...) }

func (l *Logger) Printf(template string, args ...interface{}) {
	l.logf(PrintLevel, template, args...)
}

func (l *Logger) Panicf(template string, args ...interface{}) {
	l.logf(PanicLevel, template, args...)
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.logf(FatalLevel, template, args...)
}

func (l *Logger) KVs(kv ...interface{}) *message.Record {
	e := l.newRecord()
	return e.KVs(kv...)
}

func (l *Logger) Stop() error {
	defer func() {
		l.isStop = true
	}()
	return l.engine.Stop()
}
