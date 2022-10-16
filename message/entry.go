package message

import (
	"github.com/ml444/glog/levels"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	LogName    string
	FileName   string
	FilePath   string
	CallerName string
	TradeId    string
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

const (
	maxCallerDepth int = 25
	knownFrames    int = 6
)

var (

	// qualified package name, cached at first use
	logPackage string

	// Positions in the call stack when tracing to report the calling method
	minCallerDepth int

	// Used for caller information initialisation
	callerInitOnce sync.Once
)

// getCaller retrieves the name of the first non-log calling function
func GetCaller() *runtime.Frame {
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maxCallerDepth)
		_ = runtime.Callers(0, pcs)

		// dynamic get the package name and the minimum caller depth
		for i := 0; i < maxCallerDepth; i++ {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			if strings.Contains(funcName, "getCaller") {
				logPackage = getPackageName(funcName)
				break
			}
		}

		minCallerDepth = knownFrames
	})

	// Restrict the lookback frames to avoid runaway lookups
	pcs := make([]uintptr, maxCallerDepth)
	depth := runtime.Callers(minCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)

		// If the caller isn't part of this package, we're done
		if pkg != logPackage {
			return &f //nolint:scopelint
		}
	}

	// if we got here, we failed to find the caller's context
	return nil
}

// getPackageName reduces a fully qualified function name to the package name
// There really ought to be a better way...
func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}

//func getCaller() *caller {
//	pc, file, line, ok := runtime.Caller(2)
//
//	// TODO feels nasty?
//	dir, fn := filepath.Split(file)
//	bits := strings.Split(dir, "/")
//	pkg := bits[len(bits)-2]
//
//	if ok {
//		return &caller{pc, file, line, ok, pkg, pkg, fn}
//	}
//	return nil
//}

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
