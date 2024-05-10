package formatter

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/ml444/glog/message"
)

type XMLFormatter struct {
	TimestampFormat string
	PrettyPrint     bool
}

func NewXMLFormatter(formatterCfg FormatterConfig) *XMLFormatter {
	return &XMLFormatter{
		TimestampFormat: formatterCfg.TimestampFormat,
		PrettyPrint:     formatterCfg.PrettyPrint,
	}
}

func (f *XMLFormatter) Format(entry *message.Entry) ([]byte, error) {
	record := entry.FillRecord(f.TimestampFormat)
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
