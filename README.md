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
By default, logs are output to standard output, so for production environments, we need to make the following settings (simple):
```go
package main

import (
	"os"
	
	"github.com/ml444/glog"
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/level"
)

func main() {
	var err error
	err = log.InitLog(
		config.SetLoggerName("serviceName"),
		config.SetLevel2Logger(level.DebugLevel),
		config.SetHandlerType2Logger(config.HandlerTypeFile),
		config.SetCacheSize2Logger(1024*8),
		config.SetFileDir2Logger("/var/log"),
	)
	if err != nil {
		log.Errorf("err: %v", err)
		os.Exit(-1)
	}
	// doing something
	log.Info("hello world")
	// doing something

	_ = log.Exit()
}
```
More detailed configuration can be seen in the code: `config/option.go` and `config/config.go`.
In the log storage selection with files, use the rolling way to keep the files, the default value to keep the latest 24 copies, you can adjust the number of backups according to your actual needs `SetFileBackupCount2Logger()`.
And the way of scrolling can be done by scrolling by specified size (`FileRotatorTypeTime`), scrolling by time (`FileRotatorTypeSize`), scrolling by time and size common limit (`FileRotatorTypeTimeAndSize`).
The third type of `FileRotatorTypeTimeAndSize` is described here in particular. It scrolls by time, but when it reaches the specified size limit, it stops logging and discards the rest of the log until the next point in time before a new file starts.
This is done to protect the server's disk.


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

## Report feature
In production environment sometimes we need to store some special logs, such as:
1. keep the error logs to keep them for a longer period of time. To facilitate us to trace some bugs.
2. when some high level logs need to be notified by the system alarm component, developers do not need to develop these special logging components.
3. when some special logs need special operations, such as operation logs into the database.
4. etc.

By enabling the report feature and combining it with the filter function to filter the required data and perform special operations, it makes the style and management uniform and developer-friendly.
```go
package main

import (
	"os"

	"github.com/ml444/glog"
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/level"
)

func main() {
	var err error
	err = log.InitLog(
		config.SetLoggerName("serviceName"),
		config.SetLevel2Logger(level.DebugLevel),
		config.SetHandlerType2Logger(config.HandlerTypeFile),
		config.SetCacheSize2Logger(1024*8),
		config.SetFileDir2Logger("/var/log"),

		config.SetEnableReport(),
		config.SetLevel2Logger(level.WarnLevel),
		config.SetHandlerType2Logger(config.HandlerTypeFile),
		config.SetFileRotatorType2Report(config.FileRotatorTypeSize),
		config.SetCacheSize2Logger(1024*2),
		config.SetFileDir2Logger("/var/report"),
	)
	if err != nil {
		log.Errorf("err: %v", err)
		os.Exit(-1)
	}
	// doing something
	log.Info("hello world")
	// doing something

	_ = log.Exit()
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
%[LevelNo]s         the logging level in numeric form.
%[LevelName]s       the text name of the logging level.
%[Caller]s          Execution logging call (including file path and name, line number, function name).
%[File]s            The path and file name of the source file where the logging call is executed.
%[Func]s            The name of the function called in the execution log.
%[Line]d            The line number of the execution log call.
%[DateTime]s        The time of the log execution.
%[Msecs]d           The millisecond portion of the execution logging time.
%[TradeId]s         the ID of the context trace.
%[IP]s              local IP address of the server.
%[HostName]s        Server host name.
%[Pid]d             Process ID.
%[RoutineId]d       Concurrent process ID.
%[Message]s         Message recorded.
```
**Note:**
In systems with microservices, a similar `TradeId` is typically present to assist in stringing together the entire call chain.
**glog** makes the logger automatically get the `TradeId` by configuring the hook function of `TradeIDFunc`.


