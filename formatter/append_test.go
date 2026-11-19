package formatter

import (
	"bytes"
	"runtime"
	"testing"
	"time"

	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
)

func TestAppendFormattersPreservePrefixAndOutput(t *testing.T) {
	entry := &message.Entry{
		Message:   "hello",
		TraceID:   "trace-1",
		RoutineID: 42,
		Time:      time.Date(2026, 11, 19, 12, 34, 56, 123000000, time.UTC),
		Level:     level.InfoLevel,
	}
	formatters := []struct {
		name      string
		formatter interface {
			IFormatter
			AppendFormatter
		}
	}{
		{
			name: "text",
			formatter: NewTextFormatter(TextFormatterConfig{
				BaseFormatterConfig: BaseFormatterConfig{TimeLayout: "2006-01-02 15:04:05.000"},
				PatternStyle:        "%[DateTime]s %[LevelName]s %[Message]v",
			}),
		},
		{
			name:      "json",
			formatter: NewJSONFormatter(JSONFormatterConfig{BaseFormatterConfig: BaseFormatterConfig{TimeLayout: time.RFC3339Nano}}),
		},
		{
			name:      "xml",
			formatter: NewXMLFormatter(XMLFormatterConfig{BaseFormatterConfig: BaseFormatterConfig{TimeLayout: time.RFC3339Nano}}),
		},
	}

	for _, test := range formatters {
		t.Run(test.name, func(t *testing.T) {
			want, err := test.formatter.Format(entry)
			if err != nil {
				t.Fatal(err)
			}
			prefix := []byte("prefix:")
			got, err := test.formatter.Append(append([]byte(nil), prefix...), entry)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got[:len(prefix)], prefix) {
				t.Fatalf("prefix changed: %q", got)
			}
			if !bytes.Equal(got[len(prefix):], want) {
				t.Fatalf("Append output %q, want %q", got[len(prefix):], want)
			}
		})
	}
}

func TestTextAppendMatchesLegacyPatternRenderer(t *testing.T) {
	formatter := NewTextFormatter(TextFormatterConfig{
		BaseFormatterConfig: BaseFormatterConfig{
			LoggerName:  "service",
			TimeLayout:  "2006-01-02 15:04:05.000000",
			EnableColor: true,
			ShortLevel:  true,
		},
		PatternStyle: "%[LoggerName]s %[Pid]d %[RoutineId]d %[DateTime]s %[TraceId]s %[LevelName]s %[ShortCaller]s %[Message]v",
		EnableQuote:  true,
	})
	entry := &message.Entry{
		Message:   "hello world",
		TraceID:   "trace-1",
		RoutineID: 42,
		Time:      time.Date(2026, 11, 19, 12, 34, 56, 123000000, time.UTC),
		Level:     level.WarnLevel,
		Caller: &runtime.Frame{
			Function: "example.run",
			File:     "/tmp/example.go",
			Line:     27,
		},
	}

	record := formatter.ConvertToMessage(entry)
	var legacy bytes.Buffer
	formatter.ColorRenderV2(&legacy, record)
	legacy.WriteByte('\n')
	got, err := formatter.Append(nil, entry)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, legacy.Bytes()) {
		t.Fatalf("Append output %q, legacy output %q", got, legacy.Bytes())
	}
}
