package handler

import (
	"errors"
	"fmt"
)

var (
	ErrDurableUnsupported = errors.New("durable file writes are not supported")
	ErrDurabilityUnknown  = errors.New("durable write outcome is unknown")
	ErrEmptyLogRecord     = errors.New("formatter produced an empty log record")
)

// DurabilityUnknownError means an accepted write may still complete. Retrying
// can duplicate the record. Cause is usually a context cancellation or deadline.
type DurabilityUnknownError struct{ Cause error }

func (e *DurabilityUnknownError) Error() string {
	return fmt.Sprintf("%s: %v", ErrDurabilityUnknown, e.Cause)
}
func (e *DurabilityUnknownError) Unwrap() error { return e.Cause }
func (e *DurabilityUnknownError) Is(target error) bool {
	return target == ErrDurabilityUnknown
}

// FileOperationError identifies the file and the failed persistence operation.
type FileOperationError struct {
	Path  string
	Stage string
	Err   error
}

func (e *FileOperationError) Error() string {
	return fmt.Sprintf("%s %q: %v", e.Stage, e.Path, e.Err)
}
func (e *FileOperationError) Unwrap() error { return e.Err }

func fileError(path, stage string, err error) error {
	if err == nil {
		return nil
	}
	var operation *FileOperationError
	if errors.As(err, &operation) {
		return err
	}
	return &FileOperationError{Path: path, Stage: stage, Err: err}
}
