//go:build windows && plan9
// +build windows,plan9

package handler

import (
	"io"
	"os"

	"github.com/ml444/glog/config"
	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
)

type SyslogHandler struct {
	Writer io.Writer

	formatter formatter.IFormatter
	filter    filter.IFilter
}

func NewSyslogHandler(handlerCfg *config.BaseHandlerConfig) (*SyslogHandler, error) {
	h := &SyslogHandler{
		formatter: formatter.GetNewFormatter(handlerCfg.Formatter),
		filter:    filter.GetNewFilter(handlerCfg.Filter),
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
	case level.PanicLevel:
		return fmt.Panic(msg)
	case level.FatalLevel:
		return fmt.Fatalf(msg)
	case level.ErrorLevel:
		return fmt.Fatal(msg)
	case level.WarnLevel:
		return fmt.Println(msg)
	case level.InfoLevel:
		return fmt.Println(msg)
	case level.DebugLevel:
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
