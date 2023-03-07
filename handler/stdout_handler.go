package handler

import (
	"os"

	"github.com/ml444/glog/config"
	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

type DefaultHandler struct {
	formatter formatter.IFormatter
	filter    filter.IFilter
}

func NewDefaultHandler(handlerCfg *config.BaseHandlerConfig) (*DefaultHandler, error) {
	return &DefaultHandler{
		formatter: formatter.GetNewFormatter(handlerCfg.Formatter),
		filter:    handlerCfg.Filter,
	}, nil
}

func (h *DefaultHandler) Format(record *message.Entry) ([]byte, error) {
	if h.formatter != nil {
		return h.formatter.Format(record)
	}
	return nil, nil
}

func (h *DefaultHandler) Emit(record *message.Entry) error {
	if h.filter != nil {
		if ok := h.filter.Filter(record); !ok {
			return nil
			//return errors.New(fmt.Sprintf("Filter out this msg: %v", record))
		}
	}

	msgByte, err := h.Format(record)
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
	return nil
}
