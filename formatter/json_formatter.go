package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ml444/glog/message"
)

type JSONFormatterConfig struct {
	BaseFormatterConfig
	DisableHTMLEscape bool // [json formatter] allows disabling html escaping in output.
	PrettyPrint       bool // [json|xml formatter] will indent all json logs.
	//DisableTimestamp  bool // [json formatter] allows disabling automatic timestamps in output.
}

func (c *JSONFormatterConfig) WithPrettyPrint() *JSONFormatterConfig {
	c.PrettyPrint = true
	return c
}
func (c *JSONFormatterConfig) WithDisableHTMLEscape() *JSONFormatterConfig {
	c.DisableHTMLEscape = true
	return c
}
func (c *JSONFormatterConfig) WithBaseFormatterConfig(baseCfg BaseFormatterConfig) *JSONFormatterConfig {
	c.BaseFormatterConfig = baseCfg
	return c
}

type JSONFormatter struct {
	*BaseFormatter
	disableHTMLEscape bool
	prettyPrint       bool
}

func NewJSONFormatter(cfg JSONFormatterConfig) *JSONFormatter {
	return &JSONFormatter{
		BaseFormatter:     NewBaseFormatter(cfg.BaseFormatterConfig),
		disableHTMLEscape: cfg.DisableHTMLEscape,
		prettyPrint:       cfg.PrettyPrint,
	}
}

func (f *JSONFormatter) Format(entry *message.Entry) ([]byte, error) {
	record := f.ConvertToMessage(entry)
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
