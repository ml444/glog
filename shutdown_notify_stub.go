//go:build plan9 || js || wasip1

package log

import "os"

func shutdownNotifySignals() []os.Signal {
	return nil
}
