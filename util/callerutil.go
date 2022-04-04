package util

import "runtime"

func GetCallerInfo(callerSkip int) (callerFile, callerName string, callerLine int) {
	var ok bool
	var pc uintptr
	pc, callerFile, callerLine, ok = runtime.Caller(callerSkip)
	callerName = ""
	if ok {
		callerName = runtime.FuncForPC(pc).Name()
	}
	return
}