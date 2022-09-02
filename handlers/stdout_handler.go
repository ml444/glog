package handlers

import (
	"os"

	"github.com/ml444/glog/config"
	"github.com/ml444/glog/filters"
	"github.com/ml444/glog/formatters"
	"github.com/ml444/glog/message"
)

type DefaultHandler struct {
	formatter formatters.IFormatter
	filter    filters.IFilter
}

func NewDefaultHandler(handlerCfg *config.BaseHandlerConfig) (*DefaultHandler, error) {
	formatter := formatters.GetNewFormatter(handlerCfg.Formatter)
	filter := filters.GetNewFilter(handlerCfg.Filter)
	return &DefaultHandler{
		formatter: formatter,
		filter:    filter,
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
