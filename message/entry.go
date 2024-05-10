package message

import (
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/ml444/glog/level"
)

type Entry struct {
	LogName    string
	FileName   string
	FilePath   string
	TraceID    string
	CallerName string
	ErrMsg     string
	Message    string
	RoutineID  int64
	CallerLine int
	Time       time.Time
	Level      level.LogLevel
	Caller     *runtime.Frame
}

func (e Entry) IsRecordCaller() bool {
	if e.Caller != nil || (e.CallerName != "" && e.FileName != "" && e.CallerLine > 0) {
		return true
	}
	return false
}

func ConstructFieldIndexMap() map[string]int {
	dataT := reflect.TypeOf(Entry{})
	m := map[string]int{}
	for i := 0; i < dataT.NumField(); i++ {
		field := dataT.Field(i)

		key := field.Name
		m[key] = i + 1
	}
	return m
}

func GetEntryValues(entry *Entry) []interface{} {
	dataV := reflect.ValueOf(entry)
	if dataV.Kind() == reflect.Ptr {
		dataV = dataV.Elem()
	}
	var values []interface{}
	for i := 0; i < dataV.NumField(); i++ {
		v := dataV.Field(i)
		//if v.Kind() == reflect.Map {
		//	for _, key := range v.MapKeys() {
		//		values = append(values, v.MapIndex(key))
		//	}
		//} else {
		//	values = append(values, v)
		//}
		values = append(values, v)
	}
	return values
}

func (e *Entry) FillRecord(timestampFormat string) *Record {
	record := &Record{
		Level:   e.Level.String(),
		Message: e.Message,
		ErrMsg:  e.ErrMsg,
	}

	record.Datetime = e.Time.Format(timestampFormat)
	record.Timestamp = e.Time.UnixMilli()

	if e.IsRecordCaller() {
		if e.Caller != nil {
			funcVal := e.Caller.Function
			fileVal := fmt.Sprintf("%s:%d", e.Caller.File, e.Caller.Line)
			if funcVal != "" {
				record.CallerName = funcVal
			}
			if fileVal != "" {
				record.FileName = fileVal
			}
		} else {
			record.CallerName = e.CallerName
			record.FileName = fmt.Sprintf("%s:%d", e.FileName, e.CallerLine)
		}
	}
	return record
}