package formatter

import (
	"strings"
	"time"
)

type TimeFormatter struct {
	Layout       string
	SecondLayout string
}

func NewTimeFormatter(layout string) *TimeFormatter {
	start, end := fractionBounds(layout)
	secondLayout := layout
	if start >= 0 {
		secondLayout = layout[:start] + layout[end:]
	}
	return &TimeFormatter{Layout: layout, SecondLayout: secondLayout}
}

// AppendDateTime appends a timestamp directly to dst without allocating an
// intermediate string. It is the fast path used by the text formatter.
func (tf *TimeFormatter) AppendDateTime(dst []byte, t time.Time) []byte {
	return t.AppendFormat(dst, tf.Layout)
}

func (tf *TimeFormatter) SetTimeDecimalFormat(decimal int) {
	if decimal < 0 {
		decimal = 0
	}
	if decimal > 9 {
		decimal = 9
	}
	start, end := fractionBounds(tf.Layout)
	if start < 0 {
		start = strings.LastIndex(tf.Layout, "05")
		if start < 0 {
			start = len(tf.Layout)
		} else {
			start += 2
		}
		end = start
	}
	fraction := ""
	if decimal > 0 {
		fraction = "." + strings.Repeat("0", decimal)
	}
	tf.Layout = tf.Layout[:start] + fraction + tf.Layout[end:]
	tf.SecondLayout = tf.Layout[:start] + tf.Layout[start+len(fraction):]
}

func fractionBounds(layout string) (int, int) {
	for start := 0; start+1 < len(layout); start++ {
		if layout[start] != '.' && layout[start] != ',' {
			continue
		}
		digit := layout[start+1]
		if digit != '0' && digit != '9' {
			continue
		}
		end := start + 2
		for end < len(layout) && layout[end] == digit {
			end++
		}
		if end < len(layout) && layout[end] >= '0' && layout[end] <= '9' {
			continue
		}
		return start, end
	}
	return -1, -1
}

// FormatDateTime formats a timestamp using the configured layout.
func (tf *TimeFormatter) FormatDateTime(t time.Time) string {
	return t.Format(tf.Layout)
}
