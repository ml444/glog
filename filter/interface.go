package filter

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/message"
)

type IFilter interface {
	Filter(record *message.Entry) bool
}

func GetNewFilter(filterCfg config.FilterConfig) IFilter {
	return nil
}
