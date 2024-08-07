package log

import (
	"runtime"
)

const callerSkipOffset = 6

func GetCallerFrame(callerSkip int) *runtime.Frame {
	rpc := make([]uintptr, 1)
	n := runtime.Callers(callerSkipOffset+callerSkip, rpc[:])
	if n < 1 {
		return &runtime.Frame{}
	}
	frame, _ := runtime.CallersFrames(rpc).Next()
	return &frame
}
