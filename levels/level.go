package levels

import (
	"fmt"
	"strings"
)

type LogLevel int8



const (
	DebugLevel LogLevel = iota + 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
	ReportLevel
)

// A constant exposing all logging levels
var AllLevels = []LogLevel{
	FatalLevel,
	PanicLevel,
	DPanicLevel,
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
}

// Convert the Level to a string. E.g. PanicLevel becomes "panic".
func (lvl LogLevel) String() string {
	switch lvl {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case DPanicLevel:
		return "dpanic"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	default:
		return fmt.Sprintf("Level(%d)", lvl)
	}

}

// ParseLevel takes a string levels and returns the log levels constant.
func ParseLevel(lvl string) (LogLevel, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	}

	var l LogLevel
	return l, fmt.Errorf("not a valid log Level: %q", lvl)
}


func (lvl LogLevel) ShortString() string {
	switch lvl {
	case DebugLevel:
		return "DBG "
	case InfoLevel:
		return "INF "
	case WarnLevel:
		return "WAR "
	case ErrorLevel:
		return "ERR "
	case DPanicLevel:
		return "PAN "
	case PanicLevel:
		return "PAN "
	case FatalLevel:
		return "FAT "
	default:
		return fmt.Sprintf("L(%d) ", lvl)
	}
}
//func (lvl LogLevel) Color() string {
//	switch lvl {
//	case DebugLevel, InfoLevel, ImportantLevel:
//		return green
//	case WarnLevel:
//		return yellow
//	default:
//		return red
//	}
//}
