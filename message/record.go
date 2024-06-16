package message

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/ml444/glog/level"
)

type Record struct {
	LogName      string
	ErrMsg       string
	Message      string
	TraceID      string
	RoutineID    int64
	Time         time.Time
	Level        level.LogLevel
	Caller       *runtime.Frame
	Extra        map[string]interface{}
	SendCallback func(record *Record)
}

func (x *Record) log(lvl level.LogLevel, args ...interface{}) {
	if lvl < x.Level {
		return
	}
	x.Message = fmt.Sprint(args...)
	x.SendCallback(x)
}

func (x *Record) logf(lvl level.LogLevel, template string, args ...interface{}) {
	if lvl < x.Level {
		return
	}

	msg := template
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(template, args...)
	}
	x.Message = msg
	x.SendCallback(x)
}
func (x *Record) Debug(args ...interface{}) { x.log(level.DebugLevel, args...) }
func (x *Record) Info(args ...interface{})  { x.log(level.InfoLevel, args...) }
func (x *Record) Warn(args ...interface{})  { x.log(level.WarnLevel, args...) }
func (x *Record) Error(args ...interface{}) { x.log(level.ErrorLevel, args...) }

func (x *Record) Debugf(template string, args ...interface{}) {
	x.logf(level.DebugLevel, template, args...)
}

func (x *Record) Infof(template string, args ...interface{}) {
	x.logf(level.InfoLevel, template, args...)
}

func (x *Record) Warnf(template string, args ...interface{}) {
	x.logf(level.WarnLevel, template, args...)
}

func (x *Record) Errorf(template string, args ...interface{}) {
	x.logf(level.ErrorLevel, template, args...)
}

func (x *Record) Print(args ...interface{}) { x.log(level.PrintLevel, args...) }
func (x *Record) Fatal(args ...interface{}) { x.log(level.FatalLevel, args...) }
func (x *Record) Panic(args ...interface{}) { x.log(level.PanicLevel, args...) }

func (x *Record) Println(args ...interface{}) { x.log(level.PrintLevel, args...) }
func (x *Record) Fatalln(args ...interface{}) { x.log(level.FatalLevel, args...) }
func (x *Record) Panicln(args ...interface{}) { x.log(level.PanicLevel, args...) }

func (x *Record) Printf(template string, args ...interface{}) {
	x.logf(level.PrintLevel, template, args...)
}

func (x *Record) Panicf(template string, args ...interface{}) {
	x.logf(level.PanicLevel, template, args...)
}

func (x *Record) Fatalf(template string, args ...interface{}) {
	x.logf(level.FatalLevel, template, args...)
}

// KVs  <string,any,string,any...>
func (x *Record) KVs(kv ...interface{}) *Record {
	if len(kv) < 2 {
		return x
	}
	if x.Extra == nil {
		x.Extra = map[string]interface{}{}
	}
	for i := 0; i < len(kv); i = i + 2 {
		key := kv[i]
		if keyStr, ok := key.(string); ok {
			x.Extra[keyStr] = kv[i+1]
		} else {
			x.Extra[toStr(key)] = kv[i+1]
		}
	}
	return x
}

func toStr(v interface{}) string {
	if v == nil {
		return ""
	}
	switch vv := v.(type) {
	case int:
		return strconv.FormatInt(int64(vv), 10)
	case int8:
		return strconv.FormatInt(int64(vv), 10)
	case int16:
		return strconv.FormatInt(int64(vv), 10)
	case int32:
		return strconv.FormatInt(int64(vv), 10)
	case int64:
		return strconv.FormatInt(vv, 10)
	case uint:
		return strconv.FormatUint(uint64(vv), 10)
	case uint8:
		return strconv.FormatUint(uint64(vv), 10)
	case uint16:
		return strconv.FormatUint(uint64(vv), 10)
	case uint32:
		return strconv.FormatUint(uint64(vv), 10)
	case uint64:
		return strconv.FormatUint(vv, 10)
	case string:
		return vv
	case bool:
		if vv {
			return "true"
		}
		return "false"

	}
	return fmt.Sprintf("%v", v)
}
