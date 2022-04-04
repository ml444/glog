package util

import "time"

var formatTimeSec uint32
var formatTimeSecStr string

func FormatTime(t time.Time) string {
	sec := uint32(t.Unix())
	pre := formatTimeSec
	preStr := formatTimeSecStr
	if pre == sec {
		// 受并行优化的影响，小概率取了旧值，因为是打LOG，就不搞这么严谨了
		return preStr
	}
	x := t.Format("01-02T15:04:05")
	formatTimeSec = sec
	formatTimeSecStr = x
	return x
}
