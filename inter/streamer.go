package inter

import "io"

type IStreamer interface {
	io.Writer
	Close()
}
