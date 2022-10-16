package util

import "time"

var formatTimeSec uint32
var formatTimeSecStr string

// FormatDateTime Influenced by parallelism, small probability took the old value.
// Because it is logging, we don't make it so strict
func FormatDateTime(t time.Time) string {
	sec := uint32(t.Unix())
	pre := formatTimeSec
	preStr := formatTimeSecStr
	if pre == sec {
		return preStr
	}
	x := t.Format("01-02T15:04:05")
	formatTimeSec = sec
	formatTimeSecStr = x
	return x
}
