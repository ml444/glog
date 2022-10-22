package formatter

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/message"
)

type XMLFormatter struct {
	TimestampFormat string
}

func NewXMLFormatter(formatterCfg config.FormatterConfig) *XMLFormatter {
	return &XMLFormatter{
		TimestampFormat: formatterCfg.TimestampFormat,
	}
}

func (f *XMLFormatter) Format(event *message.Entry) ([]byte, error) {
	return nil, nil
}
