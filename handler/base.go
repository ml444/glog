package handler

import (
	"github.com/ml444/glog/message"
)

type IHandler interface {
	Emit(record *message.Record) error
	Close() error
}
