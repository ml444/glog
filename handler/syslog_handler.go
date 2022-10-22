//go:build !windows && !plan9
// +build !windows,!plan9

package handler

import (
	"log/syslog"

	"github.com/ml444/glog/config"
	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
)

type SyslogHandler struct {
	//BaseHandler
	Writer   *syslog.Writer
	network  string
	raddr    string
	priority int
	tag      string

	formatter formatter.IFormatter
	filter    filter.IFilter
}

func NewSyslogHandler(handlerCfg *config.BaseHandlerConfig) (*SyslogHandler, error) {
	cfg := handlerCfg.Syslog
	h := &SyslogHandler{
		network:   cfg.Network,
		raddr:     cfg.Address,
		priority:  cfg.Priority,
		tag:       cfg.Tag,
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
	case level.PanicLevel:
		return h.Writer.Crit(msg)
	case level.FatalLevel:
		return h.Writer.Crit(msg)
	case level.ErrorLevel:
		return h.Writer.Err(msg)
	case level.WarnLevel:
		return h.Writer.Warning(msg)
	case level.InfoLevel:
		return h.Writer.Info(msg)
	case level.DebugLevel:
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
