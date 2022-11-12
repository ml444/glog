package inter

import "github.com/ml444/glog/message"

type IFilter interface {
	Filter(record *message.Entry) bool
}
