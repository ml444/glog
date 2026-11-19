//go:build darwin

package handler

import (
	"os"
	"syscall"
)

const durablePlatformSupported = true

// Do not use File.Sync here: Go may fall back to plain fsync on ENOTSUP,
// which does not provide the required device-cache flush on macOS.
func syncDurableFile(file *os.File) error {
	conn, err := file.SyscallConn()
	if err != nil {
		return err
	}
	var syncErr syscall.Errno
	err = conn.Control(func(fd uintptr) {
		for {
			_, _, syncErr = syscall.Syscall(syscall.SYS_FCNTL, fd, syscall.F_FULLFSYNC, 0)
			if syncErr != syscall.EINTR {
				break
			}
		}
	})
	if err != nil {
		return err
	}
	if syncErr != 0 {
		return syncErr
	}
	return nil
}
