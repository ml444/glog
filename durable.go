package log

import (
	"context"
	"errors"
	"fmt"

	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/handler"
)

var (
	ErrNoDurableTarget    = errors.New("no matching built-in file worker")
	ErrLoggerStopped      = errors.New("logger is stopped")
	ErrInvalidLevel       = errors.New("invalid log level")
	ErrDurableUnsupported = handler.ErrDurableUnsupported
	ErrDurabilityUnknown  = handler.ErrDurabilityUnknown
	ErrEmptyLogRecord     = handler.ErrEmptyLogRecord
	ErrLogFiltered        = filter.ErrFilterOut
)

type DurabilityUnknownError = handler.DurabilityUnknownError

// IDurableLogger is an optional extension; ILogger implementations remain valid.
type IDurableLogger interface {
	ILogger
	WriteDurable(context.Context, Level, ...interface{}) error
}

// DurableTargetError identifies a failed target by its zero-based worker index.
type DurableTargetError struct {
	Worker int
	Path   string
	Stage  string
	Err    error
}

func (e DurableTargetError) Error() string {
	return fmt.Sprintf("worker %d %s %q: %v", e.Worker, e.Stage, e.Path, e.Err)
}
func (e DurableTargetError) Unwrap() error { return e.Err }

// DurableWriteError reports every known failed or unconfirmed file target.
// Successful targets are not rolled back. Inspect Failures before retrying.
type DurableWriteError struct{ Failures []DurableTargetError }

func (e *DurableWriteError) Error() string {
	return fmt.Sprintf("durable write failed for %d target(s): %v", len(e.Failures), e.Failures)
}
func (e *DurableWriteError) Unwrap() error {
	if len(e.Failures) > 0 {
		return &e.Failures[0]
	}
	return nil
}
func (e *DurableWriteError) Is(target error) bool {
	for _, failure := range e.Failures {
		if errors.Is(failure, target) {
			return true
		}
	}
	return false
}
func (e *DurableWriteError) As(target interface{}) bool {
	for index := range e.Failures {
		if errors.As(&e.Failures[index], target) {
			return true
		}
	}
	return false
}
