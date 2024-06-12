//go:build !linux || !amd64 || noattr

package util

func UMask(_ int) int {
	return 0
}
