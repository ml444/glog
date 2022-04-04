package formatters

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/message"
)



type IFormatter interface {
	Format(*message.Entry) ([]byte, error)
}


type Fields map[string]interface{}
type fieldKey string
// FieldMap allows customization of the key names for default fields.
type FieldMap map[fieldKey]string

func (f FieldMap) resolve(key fieldKey) string {
	if k, ok := f[key]; ok {
		return k
	}

	return string(key)
}

func GetNewFormatter(formatterCfg *config.FormatterConfig) IFormatter {
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