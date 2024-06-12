//go:build linux && amd64 && !noattr

package util

import "syscall"

func UMask(mask int) int {
	oldMask := syscall.Umask(mask)
	return oldMask
}
