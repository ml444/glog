package engines

import "github.com/ml444/glog/message"

type IEngine interface {
	Init() error
	Start() error
	Stop() error
	Send(event *message.Entry)
}
