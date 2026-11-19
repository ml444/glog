package log

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/message"
)

type durableTarget struct {
	worker int
	file   *handler.FileHandler
	done   chan error
}

// SendDurable only dispatches to built-in file handlers; other handlers retain
// their ordinary asynchronous API. Admission and acknowledgements are separate
// so Stop can close admission without blocking completion of accepted writes.
func (e *ChannelEngine) SendDurable(ctx context.Context, entry *message.Entry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if entry == nil {
		return errors.New("log entry is nil")
	}
	e.sendMu.RLock()
	if atomic.LoadUint32(&e.stop) != 0 {
		e.sendMu.RUnlock()
		return ErrLoggerStopped
	}
	var targets []durableTarget
	for index, worker := range e.workers {
		if worker.fileHandler != nil && entry.Level >= worker.levelThreshold {
			targets = append(targets, durableTarget{worker: index, file: worker.fileHandler})
		}
	}
	if len(targets) == 0 {
		e.sendMu.RUnlock()
		return ErrNoDurableTarget
	}
	var failures []DurableTargetError
	accepted := 0
	for index := range targets {
		target := &targets[index]
		worker := e.workers[target.worker]
		err := ctx.Err()
		if err == nil && atomic.LoadUint32(&e.stop) != 0 {
			err = ErrLoggerStopped
		}
		if err == nil {
			snapshot := *entry
			if entry.Caller != nil {
				caller := *entry.Caller
				snapshot.Caller = &caller
			}
			done := make(chan error, 1)
			select {
			case worker.entryChan <- engineRequest{entry: &snapshot, done: done}:
				target.done = done
				accepted++
				worker.stats.AddEnqueued()
			case <-ctx.Done():
				err = ctx.Err()
			case <-worker.stopChan:
				err = ErrLoggerStopped
			}
		}
		if err != nil {
			failures = append(failures, target.failure("enqueue", err))
		}
	}
	e.sendMu.RUnlock()
	if accepted == 0 && ctx.Err() != nil {
		return ctx.Err()
	}
	for _, target := range targets {
		if target.done == nil {
			continue
		}
		// Prefer an already available result over cancellation: it is known.
		var err error
		select {
		case err = <-target.done:
		default:
			select {
			case err = <-target.done:
			case <-ctx.Done():
				err = &handler.DurabilityUnknownError{Cause: ctx.Err()}
			}
		}
		if err != nil {
			failures = append(failures, target.failure("wait", err))
		}
	}
	if len(failures) > 0 {
		return &DurableWriteError{Failures: failures}
	}
	return nil
}

func (target durableTarget) failure(stage string, err error) DurableTargetError {
	path := target.file.Path()
	var operation *handler.FileOperationError
	if errors.As(err, &operation) {
		path, stage = operation.Path, operation.Stage
	}
	return DurableTargetError{Worker: target.worker, Path: path, Stage: stage, Err: err}
}
