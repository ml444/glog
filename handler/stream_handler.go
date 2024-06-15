package handler

import (
	"errors"
	"fmt"
	"io"

	"github.com/ml444/glog/filter"
	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

const terminator = '\n'

type StreamHandlerConfig struct {
	Streamer IStreamer
}

type IStreamer interface {
	io.WriteCloser
}

type StreamHandler struct {
	stream    IStreamer
	formatter formatter.IFormatter
	filter    filter.IFilter
}

func NewStreamHandler(cfg *StreamHandlerConfig, fm formatter.IFormatter, ft filter.IFilter) (*StreamHandler, error) {
	if cfg.Streamer == nil {
		return nil, errors.New("streamer is nil")
	}
	return &StreamHandler{
		filter:    ft,
		formatter: fm,
		stream:    cfg.Streamer,
	}, nil
}

func (h *StreamHandler) format(record *message.Entry) ([]byte, error) {
	if h.formatter != nil {
		return h.formatter.Format(record)
	}
	return nil, nil
}

func (h *StreamHandler) emit(msg []byte) error {
	msg = append(msg, terminator)
	n, err := h.stream.Write(msg)
	if err != nil {
		if errors.Is(err, io.ErrShortWrite) {
			for n < len(msg) {
				var x int
				x, err = h.stream.Write(msg[n:])
				if err != nil {
					return err
				}
				n += x
			}
		}
		return err
	}
	return nil
}

func (h *StreamHandler) Emit(record *message.Entry) error {
	if h.filter != nil {
		if ok := h.filter.Filter(record); !ok {
			return fmt.Errorf("filter out this msg: %v", record)
		}
	}

	msgByte, err := h.format(record)
	if err != nil {
		return err
	}

	return h.emit(msgByte)
}

func (h *StreamHandler) Close() error {
	return h.stream.Close()
}
