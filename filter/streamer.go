package filter

import "io"

type IStreamer interface {
	io.Writer
	Close()
}
