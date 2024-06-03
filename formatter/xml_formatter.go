package formatter

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/ml444/glog/message"
)

type XMLFormatterConfig struct {
	BaseFormatterConfig
	PrettyPrint bool // [json|xml formatter] will indent all json logs.
}

type XMLFormatter struct {
	*BaseFormatter
	PrettyPrint bool
}

func NewXMLFormatter(cfg XMLFormatterConfig) *XMLFormatter {
	return &XMLFormatter{
		BaseFormatter: NewBaseFormatter(cfg.BaseFormatterConfig),
		PrettyPrint:   cfg.PrettyPrint,
	}
}

func (f *XMLFormatter) Format(entry *message.Entry) ([]byte, error) {
	record := f.ConvertToMessage(entry)
	b := &bytes.Buffer{}
	encoder := xml.NewEncoder(b)
	if f.PrettyPrint {
		encoder.Indent("", "  ")
	}
	if err := encoder.Encode(record); err != nil {
		return nil, fmt.Errorf("failed to encoding record to XML: %w", err)
	}
	return b.Bytes(), nil
}
