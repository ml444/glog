package log

import (
	"os"
	"os/signal"
	"sync"
)

// EnvNoSignalShutdown, when non-empty, disables registerShutdownOnSignals (see shutdown_signal.go).
const EnvNoSignalShutdown = "GLOG_NO_SIGNAL_SHUTDOWN"

var registerSignalShutdownOnce sync.Once

// registerShutdownOnSignals listens for termination signals and calls Stop before exiting.
// It is a best-effort flush for graceful shutdown; it does not intercept os.Exit or plain return from main.
// Disable with environment variable EnvNoSignalShutdown before process start.
func registerShutdownOnSignals() {
	registerSignalShutdownOnce.Do(func() {
		if os.Getenv(EnvNoSignalShutdown) != "" {
			return
		}
		sigs := shutdownNotifySignals()
		if len(sigs) == 0 {
			return
		}
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, sigs...)
		go func() {
			<-ch
			Stop()
			os.Exit(1)
		}()
	})
}
