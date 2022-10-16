/*

 */

package formatters

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/ml444/glog/config"
	"github.com/ml444/glog/levels"
	"github.com/ml444/glog/message"
	"github.com/ml444/glog/util"
)

// const (
//	red    = 31
//	yellow = 33
//	blue   = 36
//	gray   = 37
// )

const (
	colorRed = uint8(iota + 91)
	colorGreen
	colorYellow
	colorBlue
	colorPurple
)
const defaultBufferGrow = 128

var (
	red      = fmt.Sprintf("\x1b[%dm", colorRed)
	green    = fmt.Sprintf("\x1b[%dm", colorGreen)
	yellow   = fmt.Sprintf("\x1b[%dm", colorYellow)
	blue     = fmt.Sprintf("\x1b[%dm", colorBlue)
	cyan     = fmt.Sprintf("\x1b[%dm", 36)
	purple   = fmt.Sprintf("\x1b[%dm", colorPurple)
	colorEnd = "\x1b[0m"
)

type writeFunc func(b *bytes.Buffer, entry *message.Entry)

// TextFormatter formats logs into text
type TextFormatter struct {
	isCustomizedWritingOrder bool
	EnableQuote              bool
	EnableQuoteEmptyFields   bool
	DisableColors            bool
	TimestampFormat          string
	msgPrefix                string
	msgSuffix                string
	sepList                  []string
	sepListLen               int
	writeFuncList            []writeFunc
	writeFuncMap             map[string]writeFunc
}

func NewTextFormatter(formatterCfg config.FormatterConfig) *TextFormatter {
	var isCustom bool
	textCfg := formatterCfg.Text
	if textCfg.Pattern != "" {
		isCustom = true
	}
	textFormatter := TextFormatter{
		isCustomizedWritingOrder: isCustom,
		TimestampFormat:          formatterCfg.TimestampFormat,
		EnableQuote:              textCfg.EnableQuote,
		EnableQuoteEmptyFields:   textCfg.EnableQuoteEmptyFields,
		DisableColors:            textCfg.DisableColors,
	}
	textFormatter.init()
	textFormatter.parseWriteFuncList(textCfg.Pattern)
	return &textFormatter
}

func (f *TextFormatter) init() {
	f.writeFuncMap = map[string]writeFunc{
		"LoggerName": f.writeLogName,
		"Caller":     f.writeCaller,
		"Pid":        f.writePid,
		"RoutineId":  f.writeRoutineId,
		"Ip":         f.writeIP,
		"HostName":   f.writeHostName,
		"File":       f.writeFilepath,
		"Line":       f.writeFuncLine,
		"Func":       f.writeFuncName,
		"TradeId":    f.writeTradeId,
		"LevelName":  f.writeLogLevel,
		"LevelNo":    f.writeLogLevelNo,
		"DateTime":   f.writeDateTime,
		"Msecs":      f.writeTimeMs,
		"Message":    f.writeMessage,
	}
}

// Format renders a single log entry
func (f *TextFormatter) Format(entry *message.Entry) ([]byte, error) {
	if c := entry.Caller; c != nil {
		entry.FileName = c.File
		//entry.CallerName = c.Function
		entry.CallerLine = c.Line
		entry.FilePath, entry.CallerName = util.GetPackageName(c.Function)
	}
	b := &bytes.Buffer{}
	b.Grow(defaultBufferGrow)
	if f.isCustomizedWritingOrder {
		f.ColorRenderV2(b, entry)
	} else {
		f.ColorRender(b, entry)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TextFormatter) ColorRender(b *bytes.Buffer, entry *message.Entry) {
	f.writeLogName(b, entry)
	b.WriteByte('(')
	f.writePid(b, entry)
	b.WriteByte(',')
	f.writeRoutineId(b, entry)
	b.WriteByte(')')
	b.WriteByte(' ')
	f.writeDateTime(b, entry)
	b.WriteByte('.')
	f.writeTimeMs(b, entry)
	b.WriteByte(' ')
	b.WriteByte('<')
	f.writeTradeId(b, entry)
	b.WriteByte('>')
	b.WriteByte(' ')
	f.writeLogLevel(b, entry)
	b.WriteByte(' ')
	f.writeCaller(b, entry)
	b.WriteByte(' ')
	f.writeMessage(b, entry)
}

func (f *TextFormatter) ColorRenderV2(b *bytes.Buffer, entry *message.Entry) {
	if f.msgPrefix != "" {
		b.WriteString(f.msgPrefix)
	}
	for i, wFunc := range f.writeFuncList {
		wFunc(b, entry)
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

	// todo 特殊字符转义
	return false
}

func (f *TextFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
}

func (f *TextFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}

func Color(l levels.LogLevel) string {
	switch l {
	case levels.DebugLevel:
		return blue
	case levels.InfoLevel:
		return green
	case levels.WarnLevel:
		return yellow
	default:
		return red
	}
}

func JoinKV(key string, value interface{}) string {
	v, _ := value.(string)
	builder := strings.Builder{}
	builder.Grow(len(key) + len("=") + len(v))
	builder.WriteString(key + "=" + v)
	return builder.String()
}

// parseFormatTemp
func parseFormatTemp(src string) string {
	regexpFieldMap := message.ConstructFieldIndexMap()
	regexpReplaceFunc := func(s string) string {
		v, ok := regexpFieldMap[s]
		if !ok {
			panic(fmt.Sprintf("%s in config.Text.Pattern,But it isn't in the field of Entry", s))
		}
		return strconv.FormatInt(int64(v), 10)
	}
	var regexpPattern = regexp.MustCompile(`%\[(\w+)?\][sdfwvtq]`)
	var subRegexpPattern = regexp.MustCompile(`(\w+)?`)
	b := regexpPattern.FindAllStringSubmatch(src, -1)
	for _, b2 := range b {
		bb := subRegexpPattern.ReplaceAllStringFunc(b2[1], regexpReplaceFunc)
		ks := strings.Replace(b2[0], b2[1], bb, 1)
		src = strings.ReplaceAll(src, b2[0], ks)
	}
	return src
}

func (f *TextFormatter) writeLogName(b *bytes.Buffer, entry *message.Entry) {
	b.WriteString(entry.LogName)
}
func (f *TextFormatter) writePid(b *bytes.Buffer, _ *message.Entry) {
	b.WriteString(pidStr)
}
func (f *TextFormatter) writeIP(b *bytes.Buffer, _ *message.Entry) {
	b.WriteString(localIP)
}
func (f *TextFormatter) writeHostName(b *bytes.Buffer, _ *message.Entry) {
	b.WriteString(localHostname)
}
func (f *TextFormatter) writeRoutineId(b *bytes.Buffer, entry *message.Entry) {
	b.WriteString(strconv.FormatInt(entry.RoutineId, 10))
}
func (f *TextFormatter) writeDateTime(b *bytes.Buffer, entry *message.Entry) {
	b.WriteString(util.FormatDateTime(entry.Time))
	//b.WriteString(fmt.Sprintf(".%05d ", entry.Time.Nanosecond()/100000))
}
func (f *TextFormatter) writeTimeMs(b *bytes.Buffer, entry *message.Entry) {
	b.WriteString(fmt.Sprintf("%05d ", entry.Time.Nanosecond()/100000))
}
func (f *TextFormatter) writeTradeId(b *bytes.Buffer, entry *message.Entry) {
	if entry.TradeId != "" {
		b.WriteString(entry.TradeId)
	}
}
func (f *TextFormatter) writeLogLevel(b *bytes.Buffer, entry *message.Entry) {
	if !f.DisableColors {
		b.WriteString(Color(entry.Level))
	}
	b.WriteString(entry.Level.ShortString())
	if !f.DisableColors {
		b.WriteString(colorEnd)
	}
}
func (f *TextFormatter) writeLogLevelNo(b *bytes.Buffer, entry *message.Entry) {
	if !f.DisableColors {
		b.WriteString(Color(entry.Level))
	}
	b.WriteString(strconv.FormatUint(uint64(entry.Level), 10))
	if !f.DisableColors {
		b.WriteString(colorEnd)
	}
}
func (f *TextFormatter) writeFilepath(b *bytes.Buffer, entry *message.Entry) {
	b.WriteString(path.Join(entry.FilePath, path.Base(entry.FileName)))
}
func (f *TextFormatter) writeFuncLine(b *bytes.Buffer, entry *message.Entry) {
	b.WriteString(strconv.FormatInt(int64(entry.CallerLine), 10))
}
func (f *TextFormatter) writeFuncName(b *bytes.Buffer, entry *message.Entry) {
	b.WriteString(entry.CallerName)
}
func (f *TextFormatter) writeCaller(b *bytes.Buffer, entry *message.Entry) {
	if !f.DisableColors {
		b.WriteString(blue)
	}
	b.WriteString(path.Join(entry.FilePath, path.Base(entry.FileName)))
	b.WriteString(":")
	b.WriteString(strconv.FormatInt(int64(entry.CallerLine), 10))
	b.WriteString(":")
	b.WriteString(entry.CallerName)
	if !f.DisableColors {
		b.WriteString(colorEnd)
	}
	//b.WriteString(" ")
}
func (f *TextFormatter) writeMessage(b *bytes.Buffer, entry *message.Entry) {
	stringVal, ok := entry.Message.(string)
	if !ok {
		stringVal = fmt.Sprint(entry.Message)
	}

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
	var regexpPattern = regexp.MustCompile(`%\[(\w+)?\][sdfwvtq]`)
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
		fn, ok := f.writeFuncMap[key]
		if ok {
			list = append(list, fn)
		} else {
			println(fmt.Sprintf("%s in config.Text.Pattern,but it isn't in writeFuncMap", key))
		}
		if preIdx != 0 {
			strList = append(strList, src[preIdx:idxList[0]])
		}
		preIdx = idxList[1]
	}
	f.sepList = strList
	f.sepListLen = len(strList)
	f.writeFuncList = list
	return
}
