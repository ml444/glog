//go:build !windows && !plan9 && !js && !wasip1

package log

import (
	"os"
	"syscall"
)

func shutdownNotifySignals() []os.Signal {
	return []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
}
