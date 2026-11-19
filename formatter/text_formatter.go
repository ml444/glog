package formatter

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strconv"

	"github.com/ml444/glog/message"
)

type writeFunc func(b *bytes.Buffer, m *message.Record)
type entryWriteFunc func(dst []byte, entry *message.Entry) []byte

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
	entryFuncList []entryWriteFunc
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
func (f *TextFormatter) Format(entry *message.Entry) ([]byte, error) {
	return f.Append(make([]byte, 0, defaultBufferGrow), entry)
}

// Append renders an entry into dst without allocating an intermediate record
// or output buffer. It is used by buffered handlers to build a batch.
func (f *TextFormatter) Append(dst []byte, entry *message.Entry) ([]byte, error) {
	if f.isCustomizedWritePattern {
		dst = f.appendPattern(dst, entry)
	} else {
		dst = f.appendDefault(dst, entry)
	}
	return append(dst, '\n'), nil
}

func (f *TextFormatter) appendDefault(dst []byte, entry *message.Entry) []byte {
	dst = f.appendLogName(dst, entry)
	dst = append(dst, '(')
	dst = f.appendPid(dst, entry)
	dst = append(dst, ',')
	dst = f.appendRoutineID(dst, entry)
	dst = append(dst, ')', ' ')
	dst = f.appendDateTime(dst, entry)
	dst = append(dst, ' ', '<')
	dst = f.appendTraceID(dst, entry)
	dst = append(dst, '>', ' ')
	dst = f.appendLogLevel(dst, entry)
	dst = append(dst, ' ')
	dst = f.appendCaller(dst, entry)
	dst = append(dst, ' ')
	return f.appendMessage(dst, entry)
}

func (f *TextFormatter) appendPattern(dst []byte, entry *message.Entry) []byte {
	dst = append(dst, f.msgPrefix...)
	for i, appendField := range f.entryFuncList {
		dst = appendField(dst, entry)
		if i < f.sepListLen {
			dst = append(dst, f.sepList[i]...)
		}
	}
	return append(dst, f.msgSuffix...)
}

func (f *TextFormatter) appendLogName(dst []byte, entry *message.Entry) []byte {
	loggerName := f.entryLoggerName(entry)
	if loggerName == "" {
		return dst
	}
	if !f.DisableColors {
		dst = append(dst, purple...)
		dst = append(dst, loggerName...)
		return append(dst, colorEnd...)
	}
	return append(dst, loggerName...)
}

func (f *TextFormatter) appendPid(dst []byte, _ *message.Entry) []byte {
	return append(dst, pidStr...)
}

func (f *TextFormatter) appendIP(dst []byte, _ *message.Entry) []byte {
	return append(dst, localIP...)
}

func (f *TextFormatter) appendHostName(dst []byte, _ *message.Entry) []byte {
	return append(dst, localHostname...)
}

func (f *TextFormatter) appendRoutineID(dst []byte, entry *message.Entry) []byte {
	return strconv.AppendInt(dst, entry.RoutineID, 10)
}

func (f *TextFormatter) appendDateTime(dst []byte, entry *message.Entry) []byte {
	return f.AppendDateTime(dst, entry.Time)
}

func (f *TextFormatter) appendTraceID(dst []byte, entry *message.Entry) []byte {
	return append(dst, entry.TraceID...)
}

func (f *TextFormatter) appendLogLevel(dst []byte, entry *message.Entry) []byte {
	levelName := entry.Level.String()
	if f.cfg.ShortLevel {
		levelName = entry.Level.ShortString()
	}
	if !f.DisableColors {
		dst = append(dst, Color(entry.Level)...)
		dst = append(dst, levelName...)
		return append(dst, colorEnd...)
	}
	return append(dst, levelName...)
}

func (f *TextFormatter) appendFileName(dst []byte, entry *message.Entry) []byte {
	if entry.Caller == nil {
		return dst
	}
	return append(dst, path.Base(entry.Caller.File)...)
}

func (f *TextFormatter) appendFilepath(dst []byte, entry *message.Entry) []byte {
	if entry.Caller == nil {
		return dst
	}
	return append(dst, entry.Caller.File...)
}

func (f *TextFormatter) appendFuncLine(dst []byte, entry *message.Entry) []byte {
	if entry.Caller == nil {
		return dst
	}
	return strconv.AppendInt(dst, int64(entry.Caller.Line), 10)
}

func (f *TextFormatter) appendFuncName(dst []byte, entry *message.Entry) []byte {
	if entry.Caller == nil {
		return dst
	}
	return append(dst, entry.Caller.Function...)
}

func (f *TextFormatter) appendFullCaller(dst []byte, entry *message.Entry) []byte {
	if entry.Caller == nil || entry.Caller.File == "" {
		return dst
	}
	if !f.DisableColors {
		dst = append(dst, blue...)
	}
	dst = append(dst, entry.Caller.File...)
	dst = append(dst, ':')
	dst = strconv.AppendInt(dst, int64(entry.Caller.Line), 10)
	dst = append(dst, ':')
	dst = append(dst, entry.Caller.Function...)
	if !f.DisableColors {
		dst = append(dst, colorEnd...)
	}
	return dst
}

func (f *TextFormatter) appendCaller(dst []byte, entry *message.Entry) []byte {
	if entry.Caller == nil || entry.Caller.File == "" {
		return dst
	}
	if !f.DisableColors {
		dst = append(dst, blue...)
	}
	dst = append(dst, path.Base(entry.Caller.File)...)
	dst = append(dst, ':')
	dst = strconv.AppendInt(dst, int64(entry.Caller.Line), 10)
	dst = append(dst, ':')
	dst = append(dst, entry.Caller.Function...)
	if !f.DisableColors {
		dst = append(dst, colorEnd...)
	}
	return dst
}

func (f *TextFormatter) appendMessage(dst []byte, entry *message.Entry) []byte {
	if f.needsQuoting(entry.Message) {
		return strconv.AppendQuote(dst, entry.Message)
	}
	return append(dst, entry.Message...)
}

func (f *TextFormatter) ColorRender(b *bytes.Buffer, m *message.Record) {
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
	f.writeTraceID(b, m)
	b.WriteByte('>')
	b.WriteByte(' ')
	f.writeLogLevel(b, m)
	b.WriteByte(' ')
	f.writeCaller(b, m)
	b.WriteByte(' ')
	f.writeMessage(b, m)
}

func (f *TextFormatter) ColorRenderV2(b *bytes.Buffer, m *message.Record) {
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

func (f *TextFormatter) writeLogName(b *bytes.Buffer, m *message.Record) {
	if m.Module == "" {
		return
	}
	b.WriteString(m.Module)
}

func (f *TextFormatter) writePid(b *bytes.Buffer, _ *message.Record) {
	b.WriteString(pidStr)
}

func (f *TextFormatter) writeIP(b *bytes.Buffer, _ *message.Record) {
	b.WriteString(localIP)
}

func (f *TextFormatter) writeHostName(b *bytes.Buffer, _ *message.Record) {
	b.WriteString(localHostname)
}

func (f *TextFormatter) writeRoutineID(b *bytes.Buffer, m *message.Record) {
	b.WriteString(strconv.FormatInt(m.RoutineID, 10))
}

func (f *TextFormatter) writeDateTime(b *bytes.Buffer, m *message.Record) {
	b.WriteString(m.Datetime)
}

func (f *TextFormatter) writeTraceID(b *bytes.Buffer, m *message.Record) {
	if m.TraceID != "" {
		b.WriteString(m.TraceID)
	}
}

func (f *TextFormatter) writeLogLevel(b *bytes.Buffer, m *message.Record) {
	b.WriteString(m.Level)
}

//func (f *TextFormatter) writeLogLevelNo(b *bytes.Buffer, m *message.Record) {
//	if f.DisableColors {
//		b.WriteString(strconv.FormatUint(uint64(m.Level), 10))
//	} else {
//		b.WriteString(Color(m.Level) + strconv.FormatUint(uint64(m.Level), 10) + colorEnd)
//	}
//}

func (f *TextFormatter) writeFileName(b *bytes.Buffer, m *message.Record) {
	b.WriteString(path.Base(m.CallerPath))
}

func (f *TextFormatter) writeFilepath(b *bytes.Buffer, m *message.Record) {
	b.WriteString(m.CallerPath)
}

func (f *TextFormatter) writeFuncLine(b *bytes.Buffer, m *message.Record) {
	b.WriteString(strconv.FormatInt(int64(m.CallerLine), 10))
}

func (f *TextFormatter) writeFuncName(b *bytes.Buffer, m *message.Record) {
	b.WriteString(m.CallerName)
}

func (f *TextFormatter) writeFullCaller(b *bytes.Buffer, m *message.Record) {
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

func (f *TextFormatter) writeCaller(b *bytes.Buffer, m *message.Record) {
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

func (f *TextFormatter) writeMessage(b *bytes.Buffer, m *message.Record) {
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
		"IP":          f.writeIP,
		"HostName":    f.writeHostName,
		"CallerFile":  f.writeFileName,
		"CallerPath":  f.writeFilepath,
		"CallerLine":  f.writeFuncLine,
		"CallerName":  f.writeFuncName,
		"TraceId":     f.writeTraceID,
		"LevelName":   f.writeLogLevel,
		//"LevelNo":    f.writeLogLevelNo,
		"DateTime": f.writeDateTime,
		//"Msecs":      f.writeTimeMs,
		"Message": f.writeMessage,
	}
	entryFuncMap := map[string]entryWriteFunc{
		"LoggerName":  f.appendLogName,
		"Caller":      f.appendFullCaller,
		"ShortCaller": f.appendCaller,
		"Pid":         f.appendPid,
		"RoutineId":   f.appendRoutineID,
		"Ip":          f.appendIP,
		"IP":          f.appendIP,
		"HostName":    f.appendHostName,
		"CallerFile":  f.appendFileName,
		"CallerPath":  f.appendFilepath,
		"CallerLine":  f.appendFuncLine,
		"CallerName":  f.appendFuncName,
		"TraceId":     f.appendTraceID,
		"LevelName":   f.appendLogLevel,
		"DateTime":    f.appendDateTime,
		"Message":     f.appendMessage,
	}
	regexpPattern := regexp.MustCompile(`%\[(\w+)?\][sdfwvtq]`)
	result := regexpPattern.FindAllStringSubmatchIndex(src, -1)
	var preIdx int
	var strList []string
	var list []writeFunc
	var entryList []entryWriteFunc
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
			entryList = append(entryList, entryFuncMap[key])
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
	f.entryFuncList = entryList
}
