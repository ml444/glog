# glog
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/ml444/glog)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/ml444/glog/master/LICENSE)
[![Coverage](https://img.shields.io/badge/version-v0.1.0-blue)](https://github.com/ml444/glog/releases/tag/v0.1.0)

Glog 是一个异步的日志记录器的库，通过配置缓存大小来适应不同的高并发需求。
同时通过各种细粒度的配置来控制日志记录器的不同行为和记录方式。

## 快速开始
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
**类似的执行结果：**
![!QuickStart](https://i.imgur.com/U0xrUeQ.png)
默认情况下，日志输出到标准输出，并且日志记录器的等级设置了Info级别，所以这个示例中的Debug级别的日志并没有被输出。
不同的日志水平，通过不同的颜色标识。

### 常规的设置
默认配置下，日志是输出到标准输出的，所以在生产环境时，我们需要进行以下设置（简单的）：
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
更多详细配置可以查看代码：`config/option.go` 和 `config/config.go`。
在日志存储选择用文件时，使用滚动的方式保持文件，默认值保留最新的24份，可以根据自己的实际需求调整备份数量 `SetFileBackupCount2Logger()`。
并且滚动的方式可以通过按指定大小滚动(`FileRotatorTypeTime`)、按时间滚动(`FileRotatorTypeSize`)、按时间和大小共同限制滚动(`FileRotatorTypeTimeAndSize`)。
这里特别说明一下第三种`FileRotatorTypeTimeAndSize`，它是按时间滚动的，但是当它到达指定的大小上限后，它就停止记录日志了，会抛弃剩下的日志，直到下一个时间点的新文件开始前。
这样做的目的是保护服务器的磁盘。

### 日志等级
为了兼容标准库的logging等级，加入了print、fatal、panic三个等级:
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

## Report 特性
在生产环境有时候我们需要存储一些特殊的日志，比如：
1. 保留错误日志，使其存留的时间更长一些。以方便我们追溯一些bug。
2. 某些高等级的日志需要通过系统的告警组件通知出去的时候，开发人员不必额外去开发这些特殊的日志记录组件。
3. 一些特殊日志需要特殊操作的时候，比如操作日志存入数据库。
4. 等等。

通过启用report特性，并结合过滤器（filter）功能，筛选所需数据，进行特殊操作，使得风格和管理统一，对开发人员友好。
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
使用text模式输出日志时，可以通过配置样式模版来控制日志的展示数据及各个信息的顺序。
默认使用：`PatternTemplate1 = "%[LoggerName]s (%[Pid]d,%[RoutineId]d) %[DateTime]s.%[Msecs]d %[LevelName]s %[Caller]s %[Message]v"`
你可以自己调整其中字段的顺序以及增删，例如： `"<<%[Pid]d,%[RoutineId]d>> %[LoggerName]s %[DateTime]s %[LevelName]s %[Message]v %[Caller]s"`
你可以根据自己的喜好进行配置。
以下是所有可以配置的选项：
```
%[LoggerName]s      记录器的名称
%[LevelNo]s         数字形式的日志记录级别
%[LevelName]s       日志记录级别的文本名称
%[Caller]s          执行日志记录调用（包括文件路径及名称、行号、函数名称）
%[File]s            执行日志记录调用的源文件的路径及文件名称
%[Func]s            执行日志记录调用的函数名称
%[Line]d            执行日志记录调用的行号
%[DateTime]s        执行日志记录的时间
%[Msecs]d           执行日志记录的时间毫秒部分
%[TradeId]s         上下文追踪的ID
%[IP]s              服务器本地IP地址
%[HostName]s        服务器主机名称
%[Pid]d             进程ID
%[RoutineId]d       协程ID
%[Message]s         记录的消息
```
**注意:**
在微服务的系统中，一般都会存在类似的`TradeId`来协助串联起整个调用链路。
glog通过配置`TradeIDFunc`的钩子函数，使得logger自动获取`TradeId`。
