package log

import (
	"runtime"
)

func GetCallerFrame() *runtime.Frame {
	rpc := make([]uintptr, 1)
	n := runtime.Callers(6, rpc[:])
	if n < 1 {
		return &runtime.Frame{}
	}
	frame, _ := runtime.CallersFrames(rpc).Next()
	return &frame
}
