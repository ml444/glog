package level

import (
	"fmt"
	"strings"
)

type LogLevel int8

const (
	NoneLevel LogLevel = iota
	DebugLevel
	PrintLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

// Convert the Level to a string. E.g.
func (lvl LogLevel) String() string {
	switch lvl {
	case DebugLevel:
		return "DEBUG"
	case PrintLevel:
		return "PRINT"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	default:
		return fmt.Sprintf("Level(%d)", lvl)
	}
}

// ParseLevel takes a string level and returns the log level constant.
func ParseLevel(lvl string) (LogLevel, error) {
	switch strings.ToLower(lvl) {
	case "fatal":
		return FatalLevel, nil
	case "panic":
		return PanicLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "print":
		return PrintLevel, nil
	case "debug":
		return DebugLevel, nil
	default:
		return 0, fmt.Errorf("not a valid log Level: %q", lvl)
	}
}

func (lvl LogLevel) ShortString() string {
	switch lvl {
	case DebugLevel:
		return "DBG"
	case InfoLevel:
		return "INF"
	case WarnLevel:
		return "WAR"
	case ErrorLevel:
		return "ERR"
	case PrintLevel:
		return "PRT"
	case PanicLevel:
		return "PAN"
	case FatalLevel:
		return "FAT"
	default:
		return fmt.Sprintf("L(%d) ", lvl)
	}
}
