package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	
	"github.com/ml444/glog/message"
)

type JsonConfig struct {
	timestampFormat   string
	disableHTMLEscape bool
	prettyPrint       bool
}

func NewJsonConfig(cfg *Config) *JsonConfig {
	return &JsonConfig{
		timestampFormat:   cfg.TimestampFormat,
		disableHTMLEscape: cfg.DisableHTMLEscape,
		prettyPrint:       cfg.PrettyPrint,
	}
}

func (f *JsonConfig) Format(entry *message.Entry) ([]byte, error) {
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
