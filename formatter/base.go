package formatter

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
	"github.com/ml444/glog/util"
)

type IFormatter interface {
	Format(*message.Record) ([]byte, error)
}

var (
	loggerName    string
	localIP       string
	localHostname string
	pidStr        string
	pid           int
)

func init() {
	var err error
	pid = os.Getpid()
	pidStr = strconv.FormatInt(int64(pid), 10)
	localHostname, err = os.Hostname()
	if err != nil {
		println(err.Error())
	}
	localIP, err = util.GetFirstLocalIp()
	if err != nil {
		println(err.Error())
	}
}

func SetLoggerName(name string) {
	loggerName = name
}

type BaseFormatterConfig struct {
	// time layout string, for example: "2006-01-02 15:04:05.000"
	TimeLayout string
	// enable color rendering
	EnableColor bool
	// short level string, for example: "Err" instead of "Error"
	ShortLevel bool
	// record pid in message
	EnablePid bool
	// record ip in message
	EnableIP bool
	// record hostname in message
	EnableHostname bool
	// record timestamp(int64 ms) in message
	EnableTimestamp bool
}

func (c BaseFormatterConfig) WithTimeLayout(layout string) BaseFormatterConfig {
	c.TimeLayout = layout
	return c
}
func (c BaseFormatterConfig) WithShortLevel() BaseFormatterConfig {
	c.ShortLevel = true
	return c
}
func (c BaseFormatterConfig) WithEnableColor() BaseFormatterConfig {
	c.EnableColor = true
	return c
}
func (c BaseFormatterConfig) WithEnablePid() BaseFormatterConfig {
	c.EnablePid = true
	return c
}
func (c BaseFormatterConfig) WithEnableIP() BaseFormatterConfig {
	c.EnableIP = true
	return c
}
func (c BaseFormatterConfig) WithEnableHostname() BaseFormatterConfig {
	c.EnableHostname = true
	return c
}
func (c BaseFormatterConfig) WithEnableTimestamp() BaseFormatterConfig {
	c.EnableTimestamp = true
	return c
}

type BaseFormatter struct {
	*TimeFormatter
	cfg BaseFormatterConfig
}

func NewBaseFormatter(cfg BaseFormatterConfig) *BaseFormatter {
	return &BaseFormatter{
		TimeFormatter: NewTimeFormatter(cfg.TimeLayout),
		cfg:           cfg,
	}
}

func (b *BaseFormatter) ConvertToMessage(e *message.Record) *message.Message {
	m := &message.Message{
		RoutineID: e.RoutineID,
		Service:   loggerName,
		Level:     e.Level.String(),
		Datetime:  b.FormatDateTime(e.Time),
		TraceID:   e.TraceID,
		Message:   e.Message,
		ErrMsg:    e.ErrMsg,
	}
	if b.cfg.EnablePid {
		m.Pid = pid
	}
	if b.cfg.ShortLevel {
		m.Level = e.Level.ShortString()
	}
	if b.cfg.EnableColor {
		m.Level = Color(e.Level) + m.Level + colorEnd
		m.Service = purple + loggerName + colorEnd
	}
	if b.cfg.EnableIP {
		m.IP = localIP
	}
	if b.cfg.EnableHostname {
		m.HostName = localHostname
	}
	if b.cfg.EnableTimestamp {
		m.Timestamp = e.Time.UnixMilli()
	}
	if e.Caller != nil {
		funcVal := e.Caller.Function
		if funcVal != "" {
			m.CallerName = funcVal
		}
		//fileVal := fmt.Sprintf("%s:%d", e.Caller.File, e.Caller.Line)
		//if fileVal != "" {
		//	m.CallerPath = fileVal
		//}
		m.CallerLine = e.Caller.Line
		m.CallerPath = e.Caller.File
	}
	return m
}

// const (
//	red    = 31
//	yellow = 33
//	blue   = 36
//	gray   = 37
// )

const (
	colorRed = uint8(iota + 91)
	colorGreen
	colorYellow
	colorBlue
	colorPurple
)
const defaultBufferGrow = 128

var (
	red      = fmt.Sprintf("\x1b[%dm", colorRed)
	green    = fmt.Sprintf("\x1b[%dm", colorGreen)
	yellow   = fmt.Sprintf("\x1b[%dm", colorYellow)
	blue     = fmt.Sprintf("\x1b[%dm", colorBlue)
	cyan     = fmt.Sprintf("\x1b[%dm", 36)
	purple   = fmt.Sprintf("\x1b[%dm", colorPurple)
	colorEnd = "\x1b[0m"
)

func Color(l level.LogLevel) string {
	switch l {
	case level.DebugLevel:
		return blue
	case level.PrintLevel:
		return cyan
	case level.InfoLevel:
		return green
	case level.WarnLevel:
		return yellow
	default:
		return red
	}
}
