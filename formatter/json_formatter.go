package formatter

import (
	"fmt"
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/message"
)

type JSONFormatter struct {
	timestampFormat   string
	disableTimestamp  bool
	disableHTMLEscape bool
	prettyPrint       bool
}

func NewJSONFormatter(formatterCfg config.FormatterConfig) *JSONFormatter {
	jsonCfg := formatterCfg.Json
	return &JSONFormatter{
		timestampFormat:   formatterCfg.TimestampFormat,
		disableTimestamp:  jsonCfg.DisableTimestamp,
		disableHTMLEscape: jsonCfg.DisableHTMLEscape,
		prettyPrint:       jsonCfg.PrettyPrint,
	}
}

func (f *JSONFormatter) Format(event *message.Entry) ([]byte, error) {
	record := f.FillRecord(event)
	return record.Bytes(f.disableHTMLEscape, f.prettyPrint)
}

func (f *JSONFormatter) FillRecord(entry *message.Entry) *message.Record {
	if entry.Message == nil {
		panic("Entry.Message must be non-nil.")
	}

	record := &message.Record{
		Level:   entry.Level.String(),
		Message: entry.Message,
		ErrMsg:  entry.ErrMsg,
	}

	record.Datetime = entry.Time.Format(f.timestampFormat)
	record.Timestamp = entry.Time.UnixMilli()

	if entry.IsRecordCaller() {
		if entry.Caller != nil {
			funcVal := entry.Caller.Function
			fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
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
