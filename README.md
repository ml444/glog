# glog

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/ml444/glog)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/ml444/glog/master/LICENSE)
[![Coverage](https://img.shields.io/badge/version-v0.1.0-blue)](https://github.com/ml444/glog/releases/tag/v0.1.0)

[中文](README_zh.md)

## Development

Run `go vet` / `go test -race` from the repo root and from the `tests/` module; see [CONTRIBUTING.md](CONTRIBUTING.md) for the full checklist and refactor notes.

Glog is a library of asynchronous loggers, with configurable cache sizes to
accommodate different high concurrency requirements. It also controls the
different behaviors and logging methods of the logger through various
fine-grained configurations, including backpressure strategies and counters for
busy logging pipelines.

## Quick start

```go
package main

import (
	"github.com/ml444/glog"
)

func main() {
	defer log.Stop()
	log.Info()
	log.Debug("hello world")                            // No output by default
	log.Info("hello world")
	log.Warn("hello world")
	log.Error("hello world")
	log.Printf("%s alone could waken love!", "Love")
	log.Debugf("%s alone could waken love!", "Love")    // No output by default
	log.Infof("%s alone could waken love!", "Love")
	log.Warnf("%s alone could waken love!", "Love")
	log.Errorf("%s alone could waken love!", "Love")
}
```
**Similar result:**
![!QuickStart](https://i.imgur.com/U0xrUeQ.png)

By default, logs are output to standard output and the logger level is set to
Print. Debug messages are filtered while Print and higher levels are emitted.
Different logging levels are identified by different colors.

### General config
By default, logs are output to standard output, If you want to save the log in a file, you need to make the following settings (simple):
```go
package main

import (
	"os"

	"github.com/ml444/glog"
)

func main() {
	err := InitLogger()
	if err != nil {
		log.Errorf("err: %v", err)
		os.Exit(-1)
	}
	// doing something
	log.Info("hello world")
	// doing something
}

// InitLogger simple configuration：
func InitLogger() error {
	return log.InitLog(
		log.SetLoggerName("serviceName"),   // optional
		log.SetWorkerConfigs(log.NewDefaultTextFileWorkerConfig("./logs")),
	)
}

// InitLogger2 Simple JSON format configuration：
func InitLogger2() error {
	return log.InitLog(
		log.SetLoggerName("serviceName"),   // optional
		log.SetWorkerConfigs(log.NewDefaultJsonFileWorkerConfig("./logs")),
	)
}
```

More detailed settings：
```go
package main

import (
	"os"
	
	"github.com/ml444/glog"
)

func main() {
	err := InitLogger()
	if err != nil {
		log.Errorf("err: %v", err)
		os.Exit(-1)
	}
	// doing something
	log.Info("hello world")
	// doing something
}

// InitLogger detailed configuration：
func InitLogger() error {
	return log.InitLog(
		log.SetLoggerName("serviceName"),   // optional
		log.SetRecordCaller(0),        // enable record caller info
		log.SetWorkerConfigs(
			log.NewWorkerConfig(log.InfoLevel, 1024).SetFileHandlerConfig(
				log.NewDefaultFileHandlerConfig("logs").
					WithFileName("text_log").       // also specify a file name
					WithFileSize(1024*1024*1024).   // 1GB
					WithBackupCount(12).            // number of log files to keep
					WithBulkSize(1024*1024).        // batch write size to hard drive
					WithInterval(60*60).            // logs are cut on an hourly basis on a rolling basis
					WithRotatorType(log.FileRotatorTypeTimeAndSize),
			).SetJSONFormatterConfig(
				log.NewDefaultJSONFormatterConfig().WithBaseFormatterConfig(
					log.NewDefaultBaseFormatterConfig().
						WithEnableHostname().       // record the hostname of the server
						WithEnableTimestamp().      // record timestamp
						WithEnablePid().            // record process id
						WithEnableIP(),             // record server ip
				),
			),
		),
	)
}
```
In the log storage selection with files, use the rolling way to keep the files, the default value to keep the latest 24 copies, you can adjust the number of backups according to your actual needs `FileHandlerConfig.WithBackupCount(count int)`.
And the way of rolling can be done by rolling by specified size (`FileRotatorTypeSize`), rolling by time (`FileRotatorTypeTime`), rolling by time and size common limit (`FileRotatorTypeTimeAndSize`).
The third type of `FileRotatorTypeTimeAndSize` is described here in particular. It rolls by time, but when it reaches the specified size limit, it stops logging and discards the rest of the log until the next point in time before a new file starts.
This is done to protect the server's disk.

More detailed configuration can be seen in the code: `option.go`, `config.go`, `default.go`, and `handler/cfg.go`.

### Durable file writes

Use `WriteDurable` for a record that must be persisted before the call succeeds:

```go
import "context"

// Configure at least one built-in file worker with InitLog or New first.
if err := log.WriteDurable(context.Background(), log.ErrorLevel, "service failed"); err != nil {
    // Persistence was not confirmed. Some targets may already contain the record.
    return err
}
// A subsequent os.Exit does not need to flush this confirmed record.
```

Both `log.WriteDurable(ctx, level, args...)` and
`logger.WriteDurable(ctx, level, args...)` wait for **every built-in file worker
whose level threshold matches** to write the entire record and strongly sync it.
Each record is synced separately. These calls only dispatch to file workers;
stdout, syslog and custom handlers are not included. Even `FatalLevel` is a pure
write and does not invoke `ThrowOnLevel` or `ExitFunc`. Existing `Info`, `Error`
and other logging methods retain their asynchronous behavior and backpressure
policies. Durable requests wait for queue space instead of dropping or sampling.

- A filtered level or file filter returns `ErrLogFiltered`; no matching file
  returns `ErrNoDurableTarget`. An empty formatted record returns
  `ErrEmptyLogRecord`. None of these conditions silently succeeds.
- `DurableWriteError.Failures` identifies failed or unconfirmed workers by index,
  path and stage. `errors.Is` and `errors.As` inspect the underlying errors across
  all targets. Successful targets are not rolled back; retries may duplicate logs.
- Cancellation before admission returns the context error without writing.
  Cancellation after admission may return `ErrDurabilityUnknown` (also matching
  the context cause). Accepted records continue through the queues. A context
  cannot interrupt an operating system write or sync already in progress.
- `Stop` drains accepted requests and waits for their completion. `InitLog`
  waits for in-flight package-level calls before retiring the previous logger.
  Write/sync failures stop that file writer from appending further records;
  inspect the file and reinitialize after addressing the failure.

Overlapping file handlers during logger replacement coordinate I/O within the
process. A replacement waits for an active rotation, and an older handler
refreshes its file descriptor and size when another handler has written/rotated.

Linux uses file and directory `fsync`; macOS explicitly requires `F_FULLFSYNC`
without falling back to weaker flushing. New directories, files and rotation
renames are persisted in order. Normal logs that later trigger rotation follow
the same protocol. File creation and rotation therefore perform additional
synchronization even for ordinary logging on Linux and macOS. Other operating
systems retain ordinary logging and return `ErrDurableUnsupported` for durable
requests. The existing `ILogger`, `IEngine` and `IHandler` interfaces are unchanged;
custom loggers may implement the optional `IDurableLogger` extension.

Size rotation uses a hidden `.NAME.glog-rotation` progress file beside the log,
with atomic updates through a `.tmp` companion. It contains rotation steps,
not log records. An interrupted rotation resumes when its size rotator is
reopened; do not remove these files or change the rotation configuration before
recovery. They are removed/reused automatically as rotation completes.

The guarantee covers confirmed records within the configured retention window
on local filesystems with a single manager per file, assuming the storage stack
honors synchronization. Concurrent goroutines are supported. Network filesystems,
multiple processes sharing a file, media failure and permanent host loss are
outside this guarantee. `WriteDurable` does not make earlier asynchronous calls
durable, provide a transaction across files, or repair unconfirmed partial tails.

Process-crash tests report acknowledged IDs to a parent process and then use
`os.Exit`, an unhandled panic or `SIGKILL`, including across file rotations. To
qualify **power-loss** behavior, use an isolated Linux or macOS VM/test device
with persistent local storage: record successful IDs on an external controller,
force a hard reset without guest shutdown, remount/reboot and verify every ID in
the active file and backups. Exercise creation of nested directories, each
rotation boundary, and subsequent ordinary writes. Keep enough backups to avoid
intentional expiry. Run this for the actual deployment filesystem/storage stack;
the process-crash suite and injected sync failures do not replace that acceptance
test. No real power-loss validation is claimed by the automated unit tests.

### Backpressure strategy and counters
`glog` has two asynchronous queues:

1. the worker queue in front of each handler, configured by `WorkerConfig.SetBackpressure`.
2. the file handler buffer before disk writes, configured by `FileHandlerConfig.WithBackpressure`.

The worker queue defaults to `BackpressureStrategyBlock`, which preserves the old blocking behavior. The file handler buffer defaults to `BackpressureStrategyDrop`, which preserves the old behavior of dropping when the internal buffer is full.

Available strategies:

| Strategy | Behavior |
| --- | --- |
| `BackpressureStrategyBlock` | Wait until the queue has room. This preserves logs but can slow callers down. |
| `BackpressureStrategyDrop` | Drop immediately when the queue is full and call `OnError`/`ErrCallback`. |
| `BackpressureStrategyTimeout` | Wait up to `Timeout`, then count a timeout and report `ErrBackpressureTimeout`. |
| `BackpressureStrategySample` | When full, attempt to keep one message every `SampleRate` attempts without blocking and drop the rest. |

Example:

```go
import (
	"fmt"
	"time"
)

func InitLogger() error {
	return log.InitLog(
		log.SetLoggerName("serviceName"),
		log.SetOnError(func(v interface{}, err error) {
			// send this to your metrics or alerting system
			fmt.Printf("log backpressure: %v\n", err)
		}),
		log.SetWorkerConfigs(
			log.NewWorkerConfig(log.InfoLevel, 1024).
				SetBackpressure(
					log.NewBackpressureConfig(log.BackpressureStrategyTimeout).
						WithTimeout(50 * time.Millisecond),
				).
				SetFileHandlerConfig(
					log.NewDefaultFileHandlerConfig("./logs").
						WithBackpressure(
							log.NewBackpressureConfig(log.BackpressureStrategyDrop),
						),
				).
				SetJSONFormatterConfig(log.NewDefaultJSONFormatterConfig()),
		),
	)
}
```

Backpressure counters can be read through `log.Stats()` or `logger.Stats()`:

```go
stats := log.Stats()
for _, worker := range stats.Workers {
	fmt.Printf("queue: %+v file-buffer: %+v\n",
		worker.QueueBackpressure,
		worker.HandlerBackpressure,
	)
}
```

Each counter group contains:

- `Enqueued`: messages accepted by the queue or file buffer.
- `Dropped`: messages dropped by `Drop` or `Sample`.
- `TimedOut`: messages rejected after a timeout.
- `Sampled`: messages kept by the sample strategy while under pressure.

### High-throughput usage

Guard expensive arguments with `Enabled` so they are not computed when a level
is disabled:

```go
if log.Enabled(log.DebugLevel) {
	log.Debug("snapshot: ", buildExpensiveSnapshot())
}
```

If the output does not need goroutine IDs, disable their collection. Independent
loggers can use the same options as `InitLog`:

```go
logger, err := log.New(
	log.SetLoggerName("serviceName"),
	log.SetRecordRoutineID(false),
	log.SetWorkerConfigs(log.NewDefaultJsonFileWorkerConfig("./logs")),
)
if err != nil {
	return err
}
defer logger.Stop()
```

The file handler now applies backpressure before formatting and formats accepted
entries in reusable batches. Entries rejected by Drop or Sample do not consume
formatting work.

Time-based rotation deletes only backups exceeding `BackupCount`. Repeated time
suffixes receive numeric sequence suffixes instead of overwriting an existing
backup; these files also count toward the retention limit.

Size-limited rotators accept writes up to and including `MaxFileSize` and reject
an oversized formatted batch with `ErrFileSizeLimit`, including immediately
after a time rollover. Rejected batches are delivered to `ErrCallback` rather
than written past the limit.

The file handler snapshots the Entry and Caller before `Emit` returns, so callers
can reuse the original objects afterward. Error callbacks receive owned byte
slices that can be retained for asynchronous backups or retries. Built-in
formatters use batch appends; custom formatters always use their `Format` method
to preserve overrides such as redaction.

### Reinitialization and resource lifetime

`InitLog` publishes the new logger, waits for in-flight package-level log calls
to finish, then stops the old logger. New calls use the new logger during this
wait. Streamers and custom handlers shared through configuration are closed only
when the last logger using them stops.

`GetLogger` returns a snapshot of the current instance. Use `New` and manage
`Stop` explicitly for long-lived independent loggers. Coordinate configuration
updates in the application; do not call `InitLog` synchronously from log argument
formatters or output callbacks.

Concurrent `logger.Stop()` calls wait for the same shutdown and return its first
close error. Every handler is still closed. Once a file handler starts closing,
blocked enqueue operations are released; `Emit` after close returns
`ErrHandlerClosed`.

Worker output-error callbacks and file `ErrCallback` calls run asynchronously,
serialized per worker or file handler, independently of the output goroutines.
They may call `Stop`/`Close`. Shutdown drains accepted log entries but does not
wait for these callbacks; applications requiring callback completion must
coordinate it explicitly. Keep callbacks short to avoid queued error reports.
Producer backpressure callbacks and handler-close callbacks remain synchronous.

### Configuration inheritance and runtime levels

- Handler and formatter setters replace competing choices, including custom
  implementations; the last setter wins.
- An empty formatter `LoggerName` uses the name captured in each Entry. Runtime
  `logger.SetLoggerName` updates subsequent entries, including a reset to an
  empty name, without changing queued entries or other loggers. An explicit
  formatter name takes precedence. Use `SetLoggerName`/`GetLoggerName` instead
  of concurrently accessing `logger.Name` directly. Custom formatters can read
  `Entry.LoggerName`; `LoggerNameSet` distinguishes a captured empty name.
- `SetRecordCaller(0)` locates the calling application code for both package
  functions and logger methods. A positive skip omits application wrapper frames.
- Default workers use `NoneLevel` to follow the logger. An explicit nonzero
  worker level remains an additional threshold.
- An empty formatter `TimeLayout` inherits the global layout; an explicit layout
  takes precedence. Standalone built-in formatters use the default time layout
  when no layout is specified.
- An explicit `SetColorRender` overrides built-in text formatter color settings.
- `NewLogger` copies configuration. Failed `InitLog` calls preserve the previous
  logger and config, so retries do not inherit failed changes.
- Use `SetLevel`, `GetLevel`, and `Enabled` for concurrent runtime level access;
  do not read or write the public `Level` field concurrently.

### Enum of levels
To be compatible with the logging levels of the standard library, three levels 
of print, fatal and panic have been added.
```go
package level

type LogLevel int8
const (
	DebugLevel LogLevel = iota + 1
	PrintLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)
```

## Multi-Worker processing features
In production environment sometimes we need to store some special logs, such as:
1. keep the error logs to keep them for a longer period of time. To facilitate us to trace some bugs.
2. when some high level logs need to be notified by the system alarm component, developers do not need to develop these special logging components.
3. when some special logs need special operations, such as operation logs into the database.
4. etc.

By enabling multiple Workers and combining it with the filter function to filter the required data and perform special operations, it makes the style and management uniform and developer-friendly.
```go
package main

import (
	"os"

	"github.com/ml444/glog"
)

func main() {
	var err error
	err = InitLogger()
	if err != nil {
		log.Errorf("err: %v", err)
		os.Exit(-1)
	}
	// doing something
	log.Info("hello world")
	// doing something
}
func InitLogger() error {
	return log.InitLog(
		log.SetLoggerName("serviceName"),   // optional
		log.SetWorkerConfigs(
			log.NewDefaultStdoutWorkerConfig(),     // output to standard output
			log.NewDefaultJsonFileWorkerConfig("./logs").SetLevel(log.ErrorLevel),  // levels above error are output to a file
		),
	)
}
```

## Pattern template
When using text format mode to output logs, you can control the display data and the order of each message in the log by configuring a pattern template.
The default is: `PatternTemplateWithDefault = "%[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s %[LevelName]s %[ShortCaller]s %[Message]v"`
You can adjust the order of the fields and add or remove them yourself, for example: `"<<%[Pid]d,%[RoutineId]d>> %[LoggerName]s %[DateTime]s %[LevelName]s %[Message]v %[Caller]s"`
You can configure it to your liking.
The following are all the options that can be configured:
```
%[LoggerName]s      the name of the logger.
%[LevelName]s       the text name of the logging level.
%[ShortCaller]s     logging call (including file name, line number, function name).
%[Caller]s          logging call (including file path and name, line number, function name).
%[DateTime]s        The time of the log execution.
%[TraceId]s         the ID of the context trace.
%[IP]s              local IP address of the server.
%[HostName]s        Server host name.
%[Pid]d             Process ID.
%[RoutineId]d       Concurrent process ID.
%[Message]s         Message recorded.


%[CallerPath]s            Record the calling source file path.
%[CallerFile]s            Record the name of the source file called.
%[CallerName]s            Record the function name called.
%[CallerLine]d            Record the calling line number.
```
If you don't want to use `%[Caller]s` or `%[ShortCaller]s`, which is a fixed arrangement of caller information,
You can use `%[CallerPath]s`, `%[CallerFile]s`, `%[CallerName]s`, `%[CallerLine]d` to customize their order. For example:
```shell
%[CallerPath]s %[CallerName]s:%[CallerLine]d
%[CallerFile]s:%[CallerLine]d
```
Notes:
- `%[Caller]s` and `%[ShortCaller]s` are fixedly arranged and cannot be customized. And these two fields are mutually exclusive, only one of them can be selected.
- `%[CallerPath]s`, `%[CallerFile]s`, `%[CallerName]s`, `%[CallerLine]d` can be customized. However, the two fields `%[Caller Path]s` and `%[Caller File]s` are mutually exclusive, and only one of them can be selected.
- `%[Caller]s`, `%[ShortCaller]s` and `%[CallerPath]s`, `%[CallerFile]s`, `%[CallerName]s`, `%[CallerLine]d` are also mutually exclusive Yes, you can only choose one of the methods.

**Note:**
In systems with microservices, a similar `TraceID` is typically present to assist in stringing together the entire call chain.
**glog** makes the logger automatically get the `TraceID` by configuring the hook function of `TraceIDFunc`.
