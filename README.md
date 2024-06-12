# glog

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/ml444/glog)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/ml444/glog/master/LICENSE)
[![Coverage](https://img.shields.io/badge/version-v0.1.0-blue)](https://github.com/ml444/glog/releases/tag/v0.1.0)

[中文](README_zh.md)

Glog is a library of asynchronous loggers, with configurable cache sizes to
accommodate different high concurrency requirements. It also controls the
different behaviors and logging methods of the logger through various
fine-grained configurations.

## Quick start

```go
package main

import (
	"time"

	"github.com/ml444/glog"
)

func main() {
	log.Info()
	log.Debug("hello world")                            // No output by default
	log.Info("hello world")
	log.Warn("hello world")
	log.Error("hello world")
	time.Sleep(time.Millisecond * 10)
	log.Printf("%s alone could waken love!", "Love")    // No output by default
	log.Debugf("%s alone could waken love!", "Love")    // No output by default
	log.Infof("%s alone could waken love!", "Love")
	log.Warnf("%s alone could waken love!", "Love")
	log.Errorf("%s alone could waken love!", "Love")
}
```
**Similar result:**
![!QuickStart](https://i.imgur.com/U0xrUeQ.png)

By default, the logs are output to standard output and the logger level is set
to Info level, so the Debug level logs in this example are not output.
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
In the log storage selection with files, use the rolling way to keep the files, the default value to keep the latest 24 copies, you can adjust the number of backups according to your actual needs `SetFileBackupCount2Logger()`.
And the way of scrolling can be done by scrolling by specified size (`FileRotatorTypeTime`), scrolling by time (`FileRotatorTypeSize`), scrolling by time and size common limit (`FileRotatorTypeTimeAndSize`).
The third type of `FileRotatorTypeTimeAndSize` is described here in particular. It scrolls by time, but when it reaches the specified size limit, it stops logging and discards the rest of the log until the next point in time before a new file starts.
This is done to protect the server's disk.

More detailed configuration can be seen in the code: `config/option.go` and `config/config.go`.

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
	FatalLevel
	PanicLevel
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
The default is: `PatternTemplate1 = "%[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s.%[Msecs]d %[LevelName]s %[Caller]s %[Message]v"`
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
注意：
- `%[Caller]s` and `%[ShortCaller]s` are fixedly arranged and cannot be customized. And these two fields are mutually exclusive, only one of them can be selected.
- `%[CallerPath]s`, `%[CallerFile]s`, `%[CallerName]s`, `%[CallerLine]d` can be customized. However, the two fields `%[Caller Path]s` and `%[Caller File]s` are mutually exclusive, and only one of them can be selected.
- `%[Caller]s`, `%[ShortCaller]s` and `%[CallerPath]s`, `%[CallerFile]s`, `%[CallerName]s`, `%[CallerLine]d` are also mutually exclusive Yes, you can only choose one of the methods.

**Note:**
In systems with microservices, a similar `TraceID` is typically present to assist in stringing together the entire call chain.
**glog** makes the logger automatically get the `TraceID` by configuring the hook function of `TraceIDFunc`.


