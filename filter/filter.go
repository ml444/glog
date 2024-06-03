package filter

import (
	"errors"

	"github.com/ml444/glog/message"
)

type IFilter interface {
	Filter(record *message.Entry) bool
}

var ErrFilterOut = errors.New("filter out the message")

