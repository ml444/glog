package formatters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/message"
	"runtime"
)

type JSONFormatter struct {

	TimestampFormat string
	CallerPrettyFunc func(*runtime.Frame) (function string, file string)

	// DisableTimestamp allows disabling automatic timestamps in output
	DisableTimestamp bool

	// DisableHTMLEscape allows disabling html escaping in output
	DisableHTMLEscape bool


	// PrettyPrint will indent all json logs
	PrettyPrint bool
}

func NewJSONFormatter(formatterCfg config.FormatterConfig) *JSONFormatter {
	return &JSONFormatter{
		TimestampFormat:   formatterCfg.TimestampFormat,
		CallerPrettyFunc:  nil,
		DisableTimestamp:  false,
		DisableHTMLEscape: false,
		PrettyPrint:       false,
	}
}

func (f *JSONFormatter) Format(event *message.Entry) ([]byte, error) {
	record := f.FillRecord(event)

	b := &bytes.Buffer{}

	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(!f.DisableHTMLEscape)
	if f.PrettyPrint {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(record); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %w", err)
	}

	return b.Bytes(), nil
}
func (f *JSONFormatter) FillRecord(entry *message.Entry) *message.Record {
	if entry.Message == nil {
		panic("Entry.Message must be non-nil.")
	}

	record := &message.Record{
		Level:    entry.Level.String(),
		Message:  entry.Message,
		ErrMsg:   entry.ErrMsg,
	}

	record.Datetime = entry.Time.Format(f.TimestampFormat)
	record.Timestamp = entry.Time.UnixMilli()

	if entry.IsRecordCaller() {
		if entry.Caller != nil {
			funcVal := entry.Caller.Function
			fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
			if f.CallerPrettyFunc != nil {
				funcVal, fileVal = f.CallerPrettyFunc(entry.Caller)
			}
			if funcVal != "" {
				record.CallerName = funcVal
			}
			if fileVal != "" {
				record.FileName = fileVal
			}
		} else {
			record.CallerName = entry.CallerName
			record.FileName = fmt.Sprintf("%s:%d", entry.FileName, entry.CallerLine)
		}
	}
	return record
}
