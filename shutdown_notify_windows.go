//go:build windows

package log

import (
	"os"
	"syscall"
)

func shutdownNotifySignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}
