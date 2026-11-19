package handler

import (
	"bytes"
	"testing"

	"github.com/ml444/glog/message"
)

type bufferStream struct {
	bytes.Buffer
}

func (s *bufferStream) Close() error { return nil }

type staticFormatter struct {
	output []byte
}

func (f staticFormatter) Format(*message.Entry) ([]byte, error) {
	return append([]byte(nil), f.output...), nil
}

func TestStreamHandlerWritesExactlyOneTrailingNewline(t *testing.T) {
	for _, output := range [][]byte{[]byte("hello"), []byte("hello\n")} {
		stream := &bufferStream{}
		h, err := NewStreamHandler(
			&StreamHandlerConfig{Streamer: stream},
			staticFormatter{output: output},
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}
		if err := h.Emit(&message.Entry{}); err != nil {
			t.Fatal(err)
		}
		if got, want := stream.String(), "hello\n"; got != want {
			t.Fatalf("stream output = %q, want %q", got, want)
		}
	}
}
