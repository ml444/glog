package handler

import (
	"errors"

	"github.com/ml444/glog/message"
)

var ErrHandlerClosed = errors.New("log handler is closed")
var ErrFileSizeLimit = errors.New("maximum file size limit")

type IHandler interface {
	Emit(entry *message.Entry) error
	Close() error
}
