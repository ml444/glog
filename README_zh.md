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

#### 文件存储日志的设置
默认配置下，日志是输出到标准输出的，如果要把日志保存在文件，需要进行以下设置（简单的）：
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

// InitLogger 简单配置：
func InitLogger() error {
	return log.InitLog(
		log.SetLoggerName("serviceName"),   // 可选
		log.SetWorkerConfigs(log.NewDefaultTextFileWorkerConfig("./logs")),
	)
}

// InitLogger2 简单JSON格式配置：
func InitLogger2() error {
	return log.InitLog(
		log.SetLoggerName("serviceName"),   // 可选
		log.SetWorkerConfigs(log.NewDefaultJsonFileWorkerConfig("./logs")),
	)
}
```
更详细的设置：
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

// InitLogger 详细配置：
func InitLogger() error {
	return log.InitLog(
		log.SetLoggerName("serviceName"),   // 可选
		log.SetWorkerConfigs(
			log.NewWorkerConfig(log.InfoLevel, 1024).SetFileHandlerConfig(
                log.NewDefaultFileHandlerConfig("logs").
					WithFileName("text_log").       // 另外指定文件名
					WithFileSize(1024*1024*1024).   // 1GB
					WithBackupCount(12).            // 保留的日志文件数量
					WithBulkSize(1024*1024).        // 批量写入硬盘的大小
					WithInterval(60*60).            // 日志按每小时滚动切割
					WithRotatorType(log.FileRotatorTypeTimeAndSize),            
            ).SetJSONFormatterConfig(
                log.NewDefaultJSONFormatterConfig().WithBaseFormatterConfig(
                    log.NewDefaultBaseFormatterConfig().
                        WithEnableHostname().       // 记录服务器的hostname
                        WithEnableTimestamp().      // 记录时间戳
                        WithEnablePid().            // 记录进程ID
                        WithEnableIP(),             // 记录服务器IP
                ),
            ),
		),
	)
}
```
文件存储日志时，使用滚动的方式保持文件，默认值保留最新的24份，可以根据自己的实际需求调整备份数量 `FileHandlerConfig.WithBackupCount(count int)`。
并且滚动的方式可以通过按指定大小滚动(`FileRotatorTypeTime`)、按时间滚动(`FileRotatorTypeSize`)、按时间和大小共同限制滚动(`FileRotatorTypeTimeAndSize`)。
这里特别说明一下第三种`FileRotatorTypeTimeAndSize`，它是按时间滚动的，但是当它到达指定的大小上限后，它就停止记录日志了，会抛弃剩下的日志，直到下一个时间点的新文件开始前。
这样做的目的是保护服务器的磁盘。

更多详细配置可以查看代码：`option.go` 和 `config.go`。

### 日志等级
为了兼容标准库的logging等级，加入了print、fatal、panic三个等级:
```go
package log

type Level int8
const (
	DebugLevel Level = iota + 1
	PrintLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)
```

## 多Worker处理特性
在生产环境有时候我们需要存储一些特殊的日志，比如：
1. 保留错误日志，使其存留的时间更长一些。以方便我们追溯一些bug。
2. 某些高等级的日志需要通过系统的告警组件通知出去的时候，开发人员不必额外去开发这些特殊的日志记录组件。
3. 一些特殊日志需要特殊操作的时候，比如操作日志存入数据库。
4. 等等。

通过启用多个Worker，并结合过滤器（filter）功能，筛选所需数据，进行特殊操作，使得风格和管理统一，对开发人员友好。
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
		log.SetLoggerName("serviceName"),   // 可选
		log.SetWorkerConfigs(
			log.NewDefaultStdoutWorkerConfig(),     // 输出到标准输出
			log.NewDefaultJsonFileWorkerConfig("./logs").SetLevel(log.ErrorLevel),  // Error以上的等级输出到文件
        ),
	)
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
%[Caller]s          记录调用（包括`文件路径`、名称、行号、函数名称）
%[ShortCaller]s     记录调用，剔除文件的目录路径，使其字符串更短（只包括名称、行号、函数名称）
%[CallerPath]s      记录调用的源文件的路径
%[CallerFile]s      记录调用的源文件名称
%[CallerName]s      记录调用的函数名称
%[CallerLine]d      记录调用的行号
%[DateTime]s        记录的时间
%[TraceId]s         上下文追踪的ID
%[IP]s              服务器本地IP地址
%[HostName]s        服务器主机名称
%[Pid]d             进程ID
%[RoutineId]d       协程ID
%[Message]s         记录的消息
```
如果你不想使用`%[Caller]s`或者`%[ShortCaller]s`这个固定排列的调用者信息，
你可以使用`CallerPath`、`CallerFile`、`CallerName`、`CallerLine`来自定义他们的顺序排列。
```shell
%[CallerPath]s %[CallerName]s:%[CallerLine]d
%[CallerFile]s:%[CallerLine]d
```
注意：
- `%[Caller]s`和`%[ShortCaller]s`是固定排列的，不可自定义。且这两个字段是互斥的，只能选择其中一个。
- `%[CallerPath]s`、`%[CallerFile]s`、`%[CallerName]s`、`%[CallerLine]d`是可以自定义排列的。但`%[CallerPath]s`、`%[CallerFile]s`这两个字段是互斥的，只能选择其中一个。
- `%[Caller]s`、`%[ShortCaller]s` 和 `%[CallerPath]s`、`%[CallerFile]s`、`%[CallerName]s`、`%[CallerLine]d` 也是互斥的，只能选择其中一种方式。

**注意:**
在微服务的系统中，一般都会存在类似的`TraceID`来协助串联起整个调用链路。
glog通过配置`TraceIDFunc`的钩子函数，使得logger自动获取`TraceID`。
