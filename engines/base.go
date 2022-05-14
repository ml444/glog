package engines

import "github.com/ml444/glog/message"

type IEngine interface {
	Init() error
	Send(event *message.Entry)
	Sync() error
	Stop()
}
