//go:build windows
// +build windows

package handler

import (
	"fmt"
	"io"
	"os"

	"github.com/ml444/glog/config"
	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

type SyslogHandler struct {
	Writer io.Writer

	formatter formatter.IFormatter
	filter    filter.IFilter
}

func NewSyslogHandler(handlerCfg *config.BaseHandlerConfig) (*SyslogHandler, error) {
	h := &SyslogHandler{
		Writer:    os.Stdout,
		formatter: formatter.GetNewFormatter(handlerCfg.Formatter),
		filter:    handlerCfg.Filter,
	}
	return h, nil
}

func (h *SyslogHandler) format(record *message.Entry) ([]byte, error) {
	if h.formatter != nil {
		return h.formatter.Format(record)
	}
	return nil, nil
}

func (h *SyslogHandler) Emit(e *message.Entry) error {
	msgByte, err := h.format(e)
	if err != nil {
		return err
	}

	msg := string(msgByte)
	v := fmt.Sprintf("%s [%s] %s", e.Time.Format(config.DefaultTimestampFormat), e.Level.ShortString(), msg)
	_, err = h.Writer.Write([]byte(v))
	if err != nil {
		return err
	}
	return nil
}

func (h *SyslogHandler) Close() error {
	return nil
}
