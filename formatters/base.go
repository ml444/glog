package formatters

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/message"
	"github.com/ml444/glog/util"
	"os"
	"strconv"
)

type IFormatter interface {
	Format(*message.Entry) ([]byte, error)
}

var pidStr string
var localIP string
var localHostname string

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

func GetNewFormatter(formatterCfg config.FormatterConfig) IFormatter {
	switch formatterCfg.FormatterType {
	case config.FormatterTypeText:
		return NewTextFormatter(formatterCfg)
	case config.FormatterTypeJson:
		return NewJSONFormatter(formatterCfg)
	case config.FormatterTypeXml:
		return NewXMLFormatter(formatterCfg)
	default:
		return NewJSONFormatter(formatterCfg)
	}
}
