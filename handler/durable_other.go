//go:build !linux && !darwin

package handler

import "os"

const durablePlatformSupported = false

func syncDurableFile(*os.File) error { return ErrDurableUnsupported }
