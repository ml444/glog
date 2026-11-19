//go:build linux

package handler

import "os"

const durablePlatformSupported = true

func syncDurableFile(file *os.File) error { return file.Sync() }
