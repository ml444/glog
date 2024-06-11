package formatter

import (
	"fmt"
	"math"
	"strings"
	"time"
)

//var formatTimeSec int64
//var formatTimeSecStr string
//var formatLayout string
//var timeDecimalFormat string
//var divisor int

type TimeFormatter struct {
	SecondLayout     string
	formatTimeSecStr string
	formatTimeSec    int64
	divisor          int
}

func NewTimeFormatter(layout string) *TimeFormatter {
	secondLayout := layout
	decimal := 0
	if layout != "" {
		if idx := strings.LastIndex(layout, "."); idx > 0 {
			secondLayout = layout[:idx]
			endIdx := strings.LastIndex(layout, "Z")
			if endIdx == -1 {
				endIdx = len(layout)
			}
			decimal = len(layout[idx+1 : endIdx])
			if decimal > 9 {
				decimal = 9
			}
		}
	}
	tf := &TimeFormatter{
		SecondLayout: secondLayout,
	}
	tf.SetTimeDecimalFormat(decimal)
	return tf
}
func (tf *TimeFormatter) SetTimeDecimalFormat(decimal int) {
	if decimal <= 0 {
		return
	}
	if decimal > 9 {
		decimal = 9
	}
	if decimal == 9 {
		tf.divisor = 1
		return
	}
	tf.divisor = int(math.Pow10(9 - decimal))
}

// FormatDateTime Influenced by parallelism, small probability took the old value.
// Because it is logging, we don't make it so strict
func (tf *TimeFormatter) FormatDateTime(t time.Time) string {
	sec := t.Unix()
	pre := tf.formatTimeSec
	preStr := tf.formatTimeSecStr
	if pre == sec {
		if tf.divisor > 0 {
			return fmt.Sprintf("%s.%d", preStr, t.Nanosecond()/tf.divisor)
		}
		return preStr
	}
	x := t.Format(tf.SecondLayout)
	tf.formatTimeSec = sec
	tf.formatTimeSecStr = x
	if tf.divisor > 0 {
		return fmt.Sprintf("%s.%d", x, t.Nanosecond()/tf.divisor)
	}
	return x
}
