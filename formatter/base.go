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
type Type int

const (
	TypeText Type = 1
	TypeJson Type = 2
	TypeXml  Type = 3
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

type Config struct {
	Type              Type
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

func NewConfig(opts ...Opt) *Config {
	cfg := &Config{}
	// todo 是否需要默认值
	
	for _, opt := range opts {
		opt(cfg)
	}
	
	return cfg
}

func GetNewFormatter(formatterCfg *Config) IFormatter {
	if formatterCfg.ExternalFormatter != nil {
		return formatterCfg.ExternalFormatter
	}
	switch formatterCfg.Type {
	case TypeText:
		return NewTextConfig(formatterCfg)
	case TypeJson:
		return NewJsonConfig(formatterCfg)
	case TypeXml:
		return NewXmlConfig(formatterCfg)
	default:
		return NewJsonConfig(formatterCfg)
	}
}
