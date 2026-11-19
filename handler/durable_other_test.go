//go:build !linux && !darwin

package handler

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

func TestEmitDurableUnsupportedDoesNotWrite(t *testing.T) {
	cfg := newRegressionFileConfig(t)
	h, err := NewFileHandler(cfg, formatter.NewTextFormatter(formatter.TextFormatterConfig{PatternStyle: "%[Message]v"}), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	if err := h.EmitDurable(context.Background(), &message.Entry{Message: "unsupported"}); !errors.Is(err, ErrDurableUnsupported) {
		t.Fatalf("error = %v", err)
	}
	if _, err := os.Stat(h.Path()); !os.IsNotExist(err) {
		t.Fatalf("unsupported request opened a file: %v", err)
	}
}
