package log

import (
	"runtime"
	"strings"
)

func GetCallerFrame(callerSkip int) *runtime.Frame {
	var stack [32]uintptr
	callers := stack[:]
	count := runtime.Callers(2, callers)
	for count == len(callers) {
		callers = make([]uintptr, len(callers)*2)
		count = runtime.Callers(2, callers)
	}
	frames := runtime.CallersFrames(callers[:count])
	for count > 0 {
		frame, more := frames.Next()
		if !isLoggingFrame(frame) {
			if callerSkip <= 0 {
				return &frame
			}
			callerSkip--
		}
		if !more {
			break
		}
	}
	return &runtime.Frame{}
}

func isLoggingFrame(frame runtime.Frame) bool {
	if !strings.HasPrefix(frame.Function, "github.com/ml444/glog.") {
		return false
	}
	for _, filename := range []string{"logger.go", "log.go", "caller.go"} {
		if strings.HasSuffix(frame.File, "/"+filename) || strings.HasSuffix(frame.File, "\\"+filename) {
			return true
		}
	}
	return false
}
