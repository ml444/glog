package message

import (
	"runtime"
	"time"

	"github.com/ml444/glog/level"
)

type Entry struct {
	Message       string
	LoggerName    string
	TraceID       string
	RoutineID     int64
	Time          time.Time
	Level         level.LogLevel
	LoggerNameSet bool
	Caller        *runtime.Frame
}
