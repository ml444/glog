//go:build windows && plan9
// +build windows,plan9

package handlers

import (
	"io"
	"os"

	"github.com/ml444/glog/config"
	"github.com/ml444/glog/filters"
	"github.com/ml444/glog/formatters"
	"github.com/ml444/glog/levels"
	"github.com/ml444/glog/message"
)

type SyslogHandler struct {
	Writer io.Writer

	formatter formatters.IFormatter
	filter    filters.IFilter
}

func NewSyslogHandler(handlerCfg *config.BaseHandlerConfig) (*SyslogHandler, error) {
	formatter := formatters.GetNewFormatter(handlerCfg.Formatter)
	filter := filters.GetNewFilter(handlerCfg.Filter)
	//cfg := handlerCfg.Syslog
	h := &SyslogHandler{
		formatter: formatter,
		filter:    filter,
	}
	err := h.Init()
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (h *SyslogHandler) format(record *message.Entry) ([]byte, error) {
	if h.formatter != nil {
		return h.formatter.Format(record)
	}
	return nil, nil
}

func (h *SyslogHandler) Init() error {
	h.Writer = os.Stdout
	return nil
}

func (h *SyslogHandler) Emit(e *message.Entry) error {

	msgByte, err := h.format(e)
	if err != nil {
		return err
	}

	msg := string(msgByte)

	switch e.Level {
	case levels.PanicLevel:
		return fmt.Panic(msg)
	case levels.FatalLevel:
		return fmt.Fatalf(msg)
	case levels.ErrorLevel:
		return fmt.Fatal(msg)
	case levels.WarnLevel:
		return fmt.Println(msg)
	case levels.InfoLevel:
		return fmt.Println(msg)
	case levels.DebugLevel:
		return h.Writer.Write(msg)
	default:
		return nil
	}

}

func (h *SyslogHandler) Close() error {
	return nil
}

func (h *SyslogHandler) Flush() {

}
