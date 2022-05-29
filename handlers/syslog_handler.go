// +build !windows,!plan9

package handlers

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/filters"
	"github.com/ml444/glog/formatters"
	"github.com/ml444/glog/levels"
	"github.com/ml444/glog/message"
	"log/syslog"
)

type SyslogHandler struct {
	//BaseHandler
	Writer   *syslog.Writer
	network  string
	raddr    string
	priority int
	tag      string

	formatter formatters.IFormatter
	filter    filters.IFilter
}

func NewSyslogHandler(handlerCfg *config.BaseHandlerConfig) (*SyslogHandler, error) {
	formatter := formatters.GetNewFormatter(handlerCfg.Formatter)
	filter := filters.GetNewFilter(handlerCfg.Filter)
	cfg := handlerCfg.Syslog
	h := &SyslogHandler{
		network:   cfg.Network,
		raddr:     cfg.Address,
		priority:  cfg.Priority,
		tag:       cfg.Tag,
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
	sysLogWriter, err := syslog.Dial(h.network, h.raddr, syslog.Priority(h.priority), h.tag)
	if err != nil {
		return err
	}
	h.Writer = sysLogWriter
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

func (h *SyslogHandler) Close() error {
	return nil
}

func (h *SyslogHandler) Flush() {

}
