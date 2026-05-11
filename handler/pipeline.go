package handler

import (
	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/message"
)

// applyFilter returns filter.ErrFilterOut when the entry should not be logged.
func applyFilter(ft filter.IFilter, e *message.Entry) error {
	if ft != nil && !ft.Filter(e) {
		return filter.ErrFilterOut
	}
	return nil
}
