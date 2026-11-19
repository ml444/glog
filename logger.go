package log

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	nameMu             sync.RWMutex
	Level              Level
	levelMu            sync.RWMutex
	ThrowOnLevel       Level
	ExitFunc           func(code int) // Called after Stop on ThrowOnLevel; defaults to os.Exit (Stop is already invoked in after()).
	TraceIDFunc        func(entry *message.Entry) string
	engine             IEngine
	callerSkip         int
	enableRecordCaller bool
	enableRoutineID    bool
	isStop             uint32
}

// type FieldFunc func(entry *message.Entry) string
var (
	_ StdLogger = &Logger{}
	_ ILogger   = &Logger{}
)

// NewLogger returns a new ILogger
func NewLogger(cfg *Config) (*Logger, error) {
	if cfg == nil {
		cfg = NewDefaultConfig()
	}
	cfg = cfg.clone()
	cfg.Check()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	eng, err := NewChannelEngine(cfg)
	if err != nil {
		return nil, err
	}
	enableRoutineID := true
	if cfg.EnableRoutineID != nil {
		enableRoutineID = *cfg.EnableRoutineID
	}
	l := Logger{
		Name:               cfg.LoggerName,
		Level:              cfg.LoggerLevel,
		ThrowOnLevel:       cfg.ThrowOnLevel,
		ExitFunc:           cfg.ExitFunc,
		TraceIDFunc:        cfg.TraceIDFunc,
		engine:             eng,
		callerSkip:         cfg.CallerSkipCount,
		enableRecordCaller: cfg.EnableRecordCaller,
		enableRoutineID:    enableRoutineID,
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
	return nil
}

func (l *Logger) send(lvl Level, msg string) {
	if atomic.LoadUint32(&l.isStop) == 1 {
		println("it is stopped, can't send: ", msg)
		return
	}
	l.engine.Send(l.newEntry(lvl, msg))
}

func (l *Logger) newEntry(lvl Level, msg string) *message.Entry {
	entry := &message.Entry{
		Message:       msg,
		LoggerName:    l.GetLoggerName(),
		LoggerNameSet: true,
		Time:          time.Now(),
		Level:         lvl,
	}
	if l.enableRoutineID {
		entry.RoutineID = goid.Get()
	}
	if l.TraceIDFunc != nil {
		entry.TraceID = l.TraceIDFunc(entry)
	}

	if l.enableRecordCaller {
		entry.Caller = GetCallerFrame(l.callerSkip)
	}
	return entry
}

// WriteDurable waits until every matching built-in file worker has persisted
// the record. It never triggers ThrowOnLevel. A non-nil error does not imply
// that no bytes were written; canceled accepted writes may still complete.
func (l *Logger) WriteDurable(ctx context.Context, lvl Level, args ...interface{}) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if atomic.LoadUint32(&l.isStop) != 0 {
		return ErrLoggerStopped
	}
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return ErrDurableUnsupported
	}
	if lvl < DebugLevel || lvl > FatalLevel {
		return ErrInvalidLevel
	}
	if lvl < l.GetLevel() {
		return ErrLogFiltered
	}
	engine, ok := l.engine.(interface {
		SendDurable(context.Context, *message.Entry) error
	})
	if !ok {
		return ErrDurableUnsupported
	}
	return engine.SendDurable(ctx, l.newEntry(lvl, sprint(args)))
}

func (l *Logger) log(lvl Level, args ...interface{}) {
	if lvl < l.GetLevel() {
		return
	}
	msg := sprint(args)
	l.send(lvl, msg)
	l.after(lvl)
}

func (l *Logger) logf(lvl Level, template string, args ...interface{}) {
	if lvl < l.GetLevel() {
		return
	}

	msg := template
	if msg == "" {
		msg = sprint(args)
	} else if len(args) > 0 || strings.Contains(template, "%") {
		msg = fmt.Sprintf(template, args...)
	}
	if lvl >= PanicLevel {
		buf := new(strings.Builder)
		buf.WriteString(msg)
		l.recordStack(4, buf)
		msg = buf.String()
	}
	l.send(lvl, msg)
	l.after(lvl)
}

func sprint(args []interface{}) string {
	if len(args) == 1 {
		if messageText, ok := args[0].(string); ok {
			return messageText
		}
	}
	return fmt.Sprint(args...)
}

func (l *Logger) recordStack(callDepth int, buf *strings.Builder) {
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
		buf.WriteString(fmt.Sprintf("	[STACK]: %s %s:%d\n", name.Name(), file, line))
	}
}

func (l *Logger) after(lvl Level) {
	if lvl < PanicLevel {
		return
	}
	if l.ThrowOnLevel != NoneLevel && lvl >= l.ThrowOnLevel {
		err := l.Stop()
		if err != nil {
			println(err.Error())
		}
		l.ExitFunc(-1)
	}
}

func (l *Logger) GetLoggerName() string {
	l.nameMu.RLock()
	name := l.Name
	l.nameMu.RUnlock()
	return name
}

func (l *Logger) SetLoggerName(name string) {
	l.nameMu.Lock()
	l.Name = name
	l.nameMu.Unlock()
}

func (l *Logger) GetLevel() Level {
	l.levelMu.RLock()
	level := l.Level
	l.levelMu.RUnlock()
	return level
}

func (l *Logger) SetLevel(lvl Level) {
	l.levelMu.Lock()
	l.Level = lvl
	l.levelMu.Unlock()
}

// Enabled reports whether a level would be accepted by this logger.
func (l *Logger) Enabled(lvl Level) bool {
	return atomic.LoadUint32(&l.isStop) == 0 && lvl >= l.GetLevel()
}

// Log writes a message at the provided level.
func (l *Logger) Log(lvl Level, args ...interface{}) {
	l.log(lvl, args...)
}

// Logf writes a formatted message at the provided level.
func (l *Logger) Logf(lvl Level, template string, args ...interface{}) {
	l.logf(lvl, template, args...)
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

func (l *Logger) Stop() error {
	atomic.StoreUint32(&l.isStop, 1)
	return l.engine.Stop()
}

func (l *Logger) Stats() LoggerStats {
	if eng, ok := l.engine.(interface{ Stats() LoggerStats }); ok {
		return eng.Stats()
	}
	return LoggerStats{}
}
