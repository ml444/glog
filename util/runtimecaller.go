package util

import (
	"runtime"
	"strings"
	"sync"
)

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

// GetCaller retrieves the name of the first non-log calling function
func GetCaller() *runtime.Frame {
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maxCallerDepth)
		_ = runtime.Callers(0, pcs)

		// dynamic get the package name and the minimum caller depth
		for i := 0; i < maxCallerDepth; i++ {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			if strings.Contains(funcName, "getCaller") {
				logPackage, _ = ParsePackageName(funcName)
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
		pkg, _ := ParsePackageName(f.Function)

		// If the caller isn't part of this package, we're done
		if pkg != logPackage {
			return &f //nolint:scopelint
		}
	}

	// if we got here, we failed to find the caller's context
	return nil
}

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
