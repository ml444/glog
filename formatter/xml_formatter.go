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

func (c *XMLFormatterConfig) WithPrettyPrint() *XMLFormatterConfig {
	c.PrettyPrint = true
	return c
}
func (c *XMLFormatterConfig) WithBaseFormatterConfig(baseCfg BaseFormatterConfig) *XMLFormatterConfig {
	c.BaseFormatterConfig = baseCfg
	return c
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

func (f *XMLFormatter) Format(record *message.Record) ([]byte, error) {
	record := f.ConvertToMessage(record)
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
