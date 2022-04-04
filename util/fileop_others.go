//go:build !linux || !amd64 || noattr
// +build !linux !amd64 noattr

package util

func UMask(mask int) int {
	return 0
}
