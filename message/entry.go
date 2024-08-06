package message

import (
	"runtime"
	"time"

	"github.com/ml444/glog/level"
)

type Entry struct {
	Message   string
	TraceID   string
	RoutineID int64
	Time      time.Time
	Level     level.LogLevel
	Caller    *runtime.Frame
}
