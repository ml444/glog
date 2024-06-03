package message

import (
	"runtime"
	"time"

	"github.com/ml444/glog/level"
)

type Entry struct {
	LogName string
	//FileName   string
	//FilePath   string
	//CallerName string
	//CallerLine int
	ErrMsg    string
	Message   string
	TraceID   string
	RoutineID int64
	Time      time.Time
	Level     level.LogLevel
	Caller    *runtime.Frame
}
