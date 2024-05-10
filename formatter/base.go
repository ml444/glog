package formatter

import (
	"os"
	"strconv"

	"github.com/ml444/glog/message"
	"github.com/ml444/glog/util"
)

type IFormatter interface {
	Format(*message.Entry) ([]byte, error)
}
type FormatterType int

const (
	FormatterTypeText FormatterType = 1
	FormatterTypeJSON FormatterType = 2
	FormatterTypeXML  FormatterType = 3
)

var (
	pidStr        string
	localIP       string
	localHostname string
)

func init() {
	var err error
	pid := os.Getpid()
	pidStr = strconv.FormatInt(int64(pid), 10)
	localHostname, err = os.Hostname()
	if err != nil {
		println(err)
	}
	localIP, err = util.GetFirstLocalIp()
	if err != nil {
		println(err)
	}
}

type FormatterConfig struct {
	FormatterType     FormatterType
	ExternalFormatter IFormatter
	TimestampFormat   string

	PatternStyle           string // [text formatter] style template for formatting the data, which determines the order of the fields and the presentation style.
	EnableQuote            bool   // [text formatter] keep the string literal, while escaping safely if necessary.
	EnableQuoteEmptyFields bool   // [text formatter] when the value of field is empty, keep the string literal.
	DisableColors          bool   // [text formatter] adding color rendering to the output.

	DisableTimestamp  bool // [json formatter] allows disabling automatic timestamps in output.
	DisableHTMLEscape bool // [json formatter] allows disabling html escaping in output.
	PrettyPrint       bool // [json|xml formatter] will indent all json logs.
}

func GetNewFormatter(formatterCfg FormatterConfig) IFormatter {
	if formatterCfg.ExternalFormatter != nil {
		return formatterCfg.ExternalFormatter
	}
	switch formatterCfg.FormatterType {
	case FormatterTypeText:
		return NewTextFormatter(formatterCfg)
	case FormatterTypeJSON:
		return NewJSONFormatter(formatterCfg)
	case FormatterTypeXML:
		return NewXMLFormatter(formatterCfg)
	default:
		return NewJSONFormatter(formatterCfg)
	}
}