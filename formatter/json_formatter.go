package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ml444/glog/message"
)

type JSONFormatter struct {
	timestampFormat string
	// disableTimestamp  bool
	disableHTMLEscape bool
	prettyPrint       bool
}

func NewJSONFormatter(formatterCfg FormatterConfig) *JSONFormatter {
	return &JSONFormatter{
		timestampFormat:   formatterCfg.TimestampFormat,
		disableHTMLEscape: formatterCfg.DisableHTMLEscape,
		prettyPrint:       formatterCfg.PrettyPrint,
		// disableTimestamp:  jsonCfg.DisableTimestamp,
	}
}

func (f *JSONFormatter) Format(entry *message.Entry) ([]byte, error) {
	record := entry.FillRecord(f.timestampFormat)
	b := &bytes.Buffer{}
	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(!f.disableHTMLEscape)
	if f.prettyPrint {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(record); err != nil {
		return nil, fmt.Errorf("failed to encoding record to JSON: %w", err)
	}
	return b.Bytes(), nil
}
