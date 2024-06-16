package formatter

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strconv"

	"github.com/ml444/glog/message"
)

type writeFunc func(b *bytes.Buffer, m *message.Message)

type TextFormatterConfig struct {
	BaseFormatterConfig
	PatternStyle           string // [text formatter] style template for formatting the data, which determines the order of the fields and the presentation style.
	EnableQuote            bool   // [text formatter] keep the string literal, while escaping safely if necessary.
	EnableQuoteEmptyFields bool   // [text formatter] when the value of field is empty, keep the string literal.
}

func (c *TextFormatterConfig) WithPatternStyle(pattern string) *TextFormatterConfig {
	c.PatternStyle = pattern
	return c
}
func (c *TextFormatterConfig) WithEnableQuote() *TextFormatterConfig {
	c.EnableQuote = true
	return c
}
func (c *TextFormatterConfig) WithEnableQuoteEmptyFields() *TextFormatterConfig {
	c.EnableQuoteEmptyFields = true
	return c
}
func (c *TextFormatterConfig) WithBaseFormatterConfig(baseCfg BaseFormatterConfig) *TextFormatterConfig {
	c.BaseFormatterConfig = baseCfg
	return c
}

// TextFormatter formats logs into text
type TextFormatter struct {
	*BaseFormatter
	isCustomizedWritePattern bool
	EnableQuote              bool
	EnableQuoteEmptyFields   bool
	DisableColors            bool
	//TimestampFormat          string
	msgPrefix     string
	msgSuffix     string
	sepListLen    int
	sepList       []string
	writeFuncList []writeFunc
}

func NewTextFormatter(cfg TextFormatterConfig) *TextFormatter {
	textFormatter := TextFormatter{
		BaseFormatter: NewBaseFormatter(cfg.BaseFormatterConfig),
		//TimestampFormat:        cfg.TimestampFormat,
		EnableQuote:            cfg.EnableQuote,
		EnableQuoteEmptyFields: cfg.EnableQuoteEmptyFields,
		DisableColors:          !cfg.EnableColor,
	}
	if cfg.PatternStyle != "" {
		textFormatter.isCustomizedWritePattern = true
		textFormatter.parseWriteFuncList(cfg.PatternStyle)
	}
	return &textFormatter
}

// Format renders log's message into bytes
func (f *TextFormatter) Format(record *message.Record) ([]byte, error) {
	m := f.ConvertToMessage(record)
	b := &bytes.Buffer{}
	b.Grow(defaultBufferGrow)
	if f.isCustomizedWritePattern {
		f.ColorRenderV2(b, m)
	} else {
		f.ColorRender(b, m)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TextFormatter) ColorRender(b *bytes.Buffer, m *message.Message) {
	f.writeLogName(b, m)
	b.WriteByte('(')
	f.writePid(b, m)
	b.WriteByte(',')
	f.writeRoutineID(b, m)
	b.WriteByte(')')
	b.WriteByte(' ')
	f.writeDateTime(b, m)
	//b.WriteByte('.')
	//f.writeTimeMs(b, m)
	b.WriteByte(' ')
	b.WriteByte('<')
	f.writeTradeID(b, m)
	b.WriteByte('>')
	b.WriteByte(' ')
	f.writeLogLevel(b, m)
	b.WriteByte(' ')
	f.writeCaller(b, m)
	b.WriteByte(' ')
	f.writeMessage(b, m)
}

func (f *TextFormatter) ColorRenderV2(b *bytes.Buffer, m *message.Message) {
	if f.msgPrefix != "" {
		b.WriteString(f.msgPrefix)
	}
	for i, wFunc := range f.writeFuncList {
		wFunc(b, m)
		if i < f.sepListLen && f.sepList[i] != "" {
			b.WriteString(f.sepList[i])
		}
	}
	if f.msgSuffix != "" {
		b.WriteString(f.msgSuffix)
	}
}

func (f *TextFormatter) needsQuoting(text string) bool {
	if f.EnableQuoteEmptyFields && len(text) == 0 {
		return true
	}
	if f.EnableQuote {
		return true
	}

	return false
}

func (f *TextFormatter) writeLogName(b *bytes.Buffer, m *message.Message) {
	if m.Service == "" {
		return
	}
	b.WriteString(m.Service)
}

func (f *TextFormatter) writePid(b *bytes.Buffer, _ *message.Message) {
	b.WriteString(pidStr)
}

func (f *TextFormatter) writeIP(b *bytes.Buffer, _ *message.Message) {
	b.WriteString(localIP)
}

func (f *TextFormatter) writeHostName(b *bytes.Buffer, _ *message.Message) {
	b.WriteString(localHostname)
}

func (f *TextFormatter) writeRoutineID(b *bytes.Buffer, m *message.Message) {
	b.WriteString(strconv.FormatInt(m.RoutineID, 10))
}

func (f *TextFormatter) writeDateTime(b *bytes.Buffer, m *message.Message) {
	b.WriteString(m.Datetime)
}

func (f *TextFormatter) writeTradeID(b *bytes.Buffer, m *message.Message) {
	if m.TraceID != "" {
		b.WriteString(m.TraceID)
	}
}

func (f *TextFormatter) writeLogLevel(b *bytes.Buffer, m *message.Message) {
	b.WriteString(m.Level)
}

//func (f *TextFormatter) writeLogLevelNo(b *bytes.Buffer, m *message.Message) {
//	if f.DisableColors {
//		b.WriteString(strconv.FormatUint(uint64(m.Level), 10))
//	} else {
//		b.WriteString(Color(m.Level) + strconv.FormatUint(uint64(m.Level), 10) + colorEnd)
//	}
//}

func (f *TextFormatter) writeFileName(b *bytes.Buffer, m *message.Message) {
	b.WriteString(path.Base(m.CallerPath))
}

func (f *TextFormatter) writeFilepath(b *bytes.Buffer, m *message.Message) {
	b.WriteString(m.CallerPath)
}

func (f *TextFormatter) writeFuncLine(b *bytes.Buffer, m *message.Message) {
	b.WriteString(strconv.FormatInt(int64(m.CallerLine), 10))
}

func (f *TextFormatter) writeFuncName(b *bytes.Buffer, m *message.Message) {
	b.WriteString(m.CallerName)
}

func (f *TextFormatter) writeFullCaller(b *bytes.Buffer, m *message.Message) {
	if m.CallerPath == "" {
		return
	}
	s := m.CallerPath + ":" + strconv.FormatInt(int64(m.CallerLine), 10) + ":" + m.CallerName
	if f.DisableColors {
		b.WriteString(s)
	} else {
		b.WriteString(blue + s + colorEnd)
	}
}

func (f *TextFormatter) writeCaller(b *bytes.Buffer, m *message.Message) {
	if m.CallerPath == "" {
		return
	}
	s := path.Base(m.CallerPath) + ":" + strconv.FormatInt(int64(m.CallerLine), 10) + ":" + m.CallerName
	if f.DisableColors {
		b.WriteString(s)
	} else {
		b.WriteString(blue + s + colorEnd)
	}
}

func (f *TextFormatter) writeMessage(b *bytes.Buffer, m *message.Message) {
	stringVal := m.Message

	if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}

func (f *TextFormatter) parseWriteFuncList(src string) {
	if src == "" {
		return
	}
	writeFuncMap := map[string]writeFunc{
		"LoggerName":  f.writeLogName,
		"Caller":      f.writeFullCaller,
		"ShortCaller": f.writeCaller,
		"Pid":         f.writePid,
		"RoutineId":   f.writeRoutineID,
		"Ip":          f.writeIP,
		"HostName":    f.writeHostName,
		"CallerFile":  f.writeFileName,
		"CallerPath":  f.writeFilepath,
		"CallerLine":  f.writeFuncLine,
		"CallerName":  f.writeFuncName,
		"TradeId":     f.writeTradeID,
		"LevelName":   f.writeLogLevel,
		//"LevelNo":    f.writeLogLevelNo,
		"DateTime": f.writeDateTime,
		//"Msecs":      f.writeTimeMs,
		"Message": f.writeMessage,
	}
	regexpPattern := regexp.MustCompile(`%\[(\w+)?\][sdfwvtq]`)
	result := regexpPattern.FindAllStringSubmatchIndex(src, -1)
	var preIdx int
	var strList []string
	var list []writeFunc
	lastIdx := len(result) - 1
	for i, idxList := range result {
		if i == 0 && idxList[0] != 0 {
			f.msgPrefix = src[:idxList[0]]
		}
		if i == lastIdx && idxList[1] != len(src) {
			f.msgSuffix = src[idxList[1]:]
		}

		key := src[idxList[2]:idxList[3]]
		fn, ok := writeFuncMap[key]
		if ok {
			list = append(list, fn)
		} else {
			println(fmt.Sprintf("%s in `PatternStyle`,but it isn't in writeFuncMap", key))
		}
		if preIdx != 0 {
			strList = append(strList, src[preIdx:idxList[0]])
		}
		preIdx = idxList[1]
	}
	f.sepList = strList
	f.sepListLen = len(strList)
	f.writeFuncList = list
}
