/*
%(name)s            记录器的名称
%(levelNo)s         数字形式的日志记录级别
%(levelName)s       日志记录级别的文本名称
%(filename)s        执行日志记录调用的源文件的文件名称
%(pathname)s        执行日志记录调用的源文件的路径名称
%(funcName)s        执行日志记录调用的函数名称
%(module)s          执行日志记录调用的模块名称
%(lineno)s          执行日志记录调用的行号
%(created)s         执行日志记录的时间
%(astTime)s         日期和时间
%(msecs)s           毫秒部分
%(thread)d          线程ID
%(threadName)s      线程名称
%(process)d         进程ID
%(message)s         记录的消息
*/

package formatters

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ml444/glog/config"
	"github.com/ml444/glog/levels"
	"github.com/ml444/glog/message"
	"github.com/ml444/glog/util"
	"github.com/petermattis/goid"
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

var (
	colorEnd string
	red      string
	green    string
	yellow   string
	purple   string
	blue     string
)

func init() {
	// baseTimestamp = time.Now()
	red = fmt.Sprintf("\x1b[%dm", colorRed)
	green = fmt.Sprintf("\x1b[%dm", colorGreen)
	yellow = fmt.Sprintf("\x1b[%dm", colorYellow)
	// 这里先这样了，后面统一改颜色
	blue = fmt.Sprintf("\u001B[%d;1m", 36)
	purple = fmt.Sprintf("\x1b[%dm", colorPurple)
	colorEnd = "\x1b[0m"
}

// TextFormatter formats logs into text
type TextFormatter struct {
	strTemplate string

	TimestampFormat        string
	EnableQuote            bool
	EnableQuoteEmptyFields bool
	DisableColors          bool
}

func NewTextFormatter(formatterCfg config.FormatterConfig) *TextFormatter {
	return &TextFormatter{
		strTemplate:            ParseStrTemp(formatterCfg.Text.Pattern),
		TimestampFormat:        formatterCfg.TimestampFormat,
		EnableQuote:            formatterCfg.Text.EnableQuote,
		EnableQuoteEmptyFields: formatterCfg.Text.EnableQuoteEmptyFields,
		DisableColors:          formatterCfg.Text.DisableColors,
	}
}

func (f *TextFormatter) init() {

}

// Format renders a single log entry
func (f *TextFormatter) Format(entry *message.Entry) ([]byte, error) {
	// record := f.FillRecord(entry)

	b := &bytes.Buffer{}

	// timestampFormat := f.TimestampFormat
	// if timestampFormat == "" {
	//	timestampFormat = defaultTimestampFormat
	// }
	if !f.DisableColors {
		f.ColorRender(b, entry)
	} else {
		recordT := reflect.TypeOf(entry)
		if recordT.Kind() == reflect.Ptr {
			recordT = recordT.Elem()
		}
		recordV := reflect.ValueOf(entry)
		if recordV.Kind() == reflect.Ptr {
			recordV = recordV.Elem()
		}
		for i := 0; i < recordV.NumField(); i++ {
			Tv := recordT.Field(i)
			f.appendKeyValue(b, Tv.Name, recordV.Field(i).Interface())
		}
	}

	// b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TextFormatter) ColorRender(b *bytes.Buffer, entry *message.Entry) {
	// mod
	b.WriteString(entry.LogName)

	routineId := goid.Get()

	// 进程、协程
	b.WriteString(fmt.Sprintf("(%d,%d) ", entry.Pid, routineId))

	// 时间
	b.WriteString(util.FormatTime(entry.Time))
	b.WriteString(fmt.Sprintf(".%04d ", entry.Time.Nanosecond()/100000))

	if entry.TradeId != "" {
		b.WriteString("<")
		b.WriteString(entry.TradeId)
		b.WriteString("> ")
	}

	// 日志级别
	b.WriteString(Color(entry.Level))
	b.WriteString(entry.Level.ShortString())

	var callerName, callerFile string
	var callerLine int
	if c := entry.Caller; c != nil {
		callerFile = c.File
		callerName = c.Function
		callerLine = c.Line
	} else {
		callerFile = entry.FileName
		callerName = entry.CallerName
		callerLine = entry.CallerLine
	}

	// 调用位置
	filePath, fileFunc := util.GetPackageName(callerName)
	b.WriteString(path.Join(filePath, path.Base(callerFile)))
	b.WriteString(":")
	b.WriteString(fmt.Sprintf("%d:", callerLine))
	b.WriteString(fileFunc)
	b.WriteString(colorEnd)
	b.WriteString(" ")

	// 文本内容
	f.appendValue(b, entry.Message)
	// b.WriteString()
	b.WriteString("\n")

}

func Color(l levels.LogLevel) string {
	switch l {
	case levels.DebugLevel, levels.InfoLevel:
		return green
	case levels.WarnLevel:
		return yellow
	default:
		return red
	}
}

func Red(s string) string {
	return red + s + colorEnd
}
func Green(s string) string {
	return green + s + colorEnd
}
func Yellow(s string) string {
	return yellow + s + colorEnd
}
func Blue(s string) string {
	return blue + s + colorEnd
}
func Purple(s string) string {
	return purple + s + colorEnd
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

func JoinKV(key string, value interface{}) string {
	v, _ := value.(string)
	builder := strings.Builder{}
	builder.Grow(len(key) + len("=") + len(v))
	builder.WriteString(key + "=" + v)
	return builder.String()
}

// ParseStrTemp
func ParseStrTemp(src string) string {
	regexpFieldMap := message.ConstructFieldIndexMap()
	regexpReplaceFunc := func(s string) string {
		v, ok := regexpFieldMap[s]
		if !ok {
			panic(fmt.Sprintf("%s in config.Text.Pattern,But it isn't in the field of Entry", s))
		}
		return strconv.FormatInt(int64(v), 10)
	}
	regexpPattern := regexp.MustCompile(`%\[(\w+)?\][sdfvtq]`)
	subRegexpPattern := regexp.MustCompile(`(\w+)?`)
	b := regexpPattern.FindAllStringSubmatch(src, -1)
	for _, b2 := range b {
		bb := subRegexpPattern.ReplaceAllStringFunc(b2[1], regexpReplaceFunc)
		ks := strings.Replace(b2[0], b2[1], bb, 1)
		src = strings.ReplaceAll(src, b2[0], ks)
	}
	return src
}
