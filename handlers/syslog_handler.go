package handlers

import (
	"fmt"
	"github.com/ml444/glog/filters"
	"github.com/ml444/glog/formatters"
	"github.com/ml444/glog/levels"
	"github.com/ml444/glog/message"
	"log/syslog"
	"os"
)

type SyslogHandler struct {
	//BaseHandler
	Writer        *syslog.Writer
	SyslogNetwork string
	SyslogRaddr   string

	formatter formatters.IFormatter
	filter    filters.IFilter
}

func NewSyslogHandler(formatter formatters.IFormatter, filter filters.IFilter) (*SyslogHandler, error) {
	return &SyslogHandler{
		Writer:        nil,
		SyslogNetwork: "",
		SyslogRaddr:   "",
		formatter:     formatter,
		filter:        filter,
	}, nil
}

func (h *SyslogHandler) format(record *message.Entry) ([]byte, error) {
	if h.formatter != nil {
		return h.formatter.Format(record)
	}
	return nil, nil
}

func (h *SyslogHandler) Init(dir, name string) error {
	return nil
}

func (h *SyslogHandler) Emit(e *message.Entry) error {

	msgByte, err := h.format(e)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}

	msg := string(msgByte)

	switch e.Level {
	case levels.PanicLevel:
		return h.Writer.Crit(msg)
	case levels.FatalLevel:
		return h.Writer.Crit(msg)
	case levels.ErrorLevel:
		return h.Writer.Err(msg)
	case levels.WarnLevel:
		return h.Writer.Warning(msg)
	case levels.InfoLevel:
		return h.Writer.Info(msg)
	case levels.DebugLevel:
		return h.Writer.Debug(msg)
	default:
		return nil
	}

}

//func (h *SyslogHandler) Emit(msgByte []byte) error  {
//	return nil
//}

func (h *SyslogHandler) Sync() error {
	return nil
}

func (h *SyslogHandler) Flush() {

}
