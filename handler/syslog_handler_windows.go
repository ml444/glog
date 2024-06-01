//go:build windows
// +build windows

package handler

import (
	"fmt"
	"io"
	"os"
	"time"
	
	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

type SyslogHandler struct {
	Writer io.Writer
	
	formatter formatter.IFormatter
	filter    filter.IFilter
}

func NewSyslogHandler(handlerCfg *Config) (*SyslogHandler, error) {
	h := &SyslogHandler{
		Writer:    os.Stdout,
		formatter: formatter.GetNewFormatter(handlerCfg.FormatConfig),
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
	if h.filter != nil {
		if ok := h.filter.Filter(e); !ok {
			return filter.ErrFilterOut
		}
	}
	msgByte, err := h.format(e)
	if err != nil {
		return err
	}
	
	msg := string(msgByte)
	v := fmt.Sprintf("%s [%s] %s", e.Time.Format(time.RFC3339), e.Level.ShortString(), msg)
	_, err = h.Writer.Write([]byte(v))
	if err != nil {
		return err
	}
	return nil
}

func (h *SyslogHandler) Close() error {
	return nil
}
