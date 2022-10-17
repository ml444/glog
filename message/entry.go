package message

import (
	"github.com/ml444/glog/levels"
	"reflect"
	"runtime"
	"time"
)

type Entry struct {
	LogName    string
	FileName   string
	FilePath   string
	TradeId    string
	CallerName string
	CallerLine int
	RoutineId  int64

	Level  levels.LogLevel
	Time   time.Time
	Caller *runtime.Frame

	Message interface{}
	ErrMsg  string
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
