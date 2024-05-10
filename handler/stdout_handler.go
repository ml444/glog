package handler

import (
	"os"
	"time"

	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

type DefaultHandler struct {
	formatter formatter.IFormatter
	filter    filter.IFilter
}

func NewDefaultHandler(handlerCfg *HandlerConfig) (*DefaultHandler, error) {
	return &DefaultHandler{
		formatter: formatter.GetNewFormatter(handlerCfg.Formatter),
		filter:    handlerCfg.Filter,
	}, nil
}

func (h *DefaultHandler) Format(entry *message.Entry) ([]byte, error) {
	if h.formatter != nil {
		return h.formatter.Format(entry)
	}
	return nil, nil
}

func (h *DefaultHandler) Emit(entry *message.Entry) error {
	if h.filter != nil {
		if ok := h.filter.Filter(entry); !ok {
			return filter.ErrFilterOut
		}
	}

	msgByte, err := h.Format(entry)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(msgByte)
	if err != nil {
		return err
	}
	return nil
}

func (h *DefaultHandler) Close() error {
	<-time.After(time.Millisecond * 100)
	return nil
}