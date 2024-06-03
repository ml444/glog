package handler

import (
	"os"
	"time"

	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

type StdoutHandler struct {
	formatter formatter.IFormatter
	filter    filter.IFilter
}

func NewStdoutHandler(fm formatter.IFormatter, ft filter.IFilter) (*StdoutHandler, error) {
	return &StdoutHandler{
		formatter: fm,
		filter:    ft,
	}, nil
}

func (h *StdoutHandler) Format(entry *message.Entry) ([]byte, error) {
	if h.formatter != nil {
		return h.formatter.Format(entry)
	}
	return nil, nil
}

func (h *StdoutHandler) Emit(entry *message.Entry) error {
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

func (h *StdoutHandler) Close() error {
	<-time.After(time.Millisecond * 10)
	return nil
}
