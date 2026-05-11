package handler

import (
	"sync"
	"time"
)

var timeoutTimerPool = sync.Pool{
	New: func() interface{} {
		return time.NewTimer(time.Hour)
	},
}

// AcquireTimeoutTimer returns a pooled timer reset to d for select-based timeout backpressure.
func AcquireTimeoutTimer(d time.Duration) *time.Timer {
	t := timeoutTimerPool.Get().(*time.Timer)
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(d)
	return t
}

// ReleaseTimeoutTimer returns a timer to the pool after stopping it.
func ReleaseTimeoutTimer(t *time.Timer) {
	if t == nil {
		return
	}
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	timeoutTimerPool.Put(t)
}
