package engine

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/message"
)

type IEngine interface {
	Init() error
	Start() error
	Stop() error
	Send(event *message.Entry)
}

func NewEngine(typ config.EngineType) IEngine {
	switch typ {
	case config.EngineTypeChannel:
		return NewChannelEngine()
	case config.EngineTypeRingBuffer:
		return NewRingBufferEngine()
	default:
		return NewChannelEngine()
	}
}
