package formatter

import (
	"bytes"
	"testing"
	"time"
)

func TestTimeFormatterPreservesFullLayout(t *testing.T) {
	for _, zone := range []*time.Location{time.UTC, time.FixedZone("east", 8*3600)} {
		for _, nanos := range []int{0, 7000000, 123456789} {
			for _, layout := range []string{
				"2006-01-02 15:04:05.000",
				"2006-01-02 15:04:05,000000",
				time.RFC3339Nano,
				"2006-01-02T15:04:05.000000000-07:00",
				"2006.01.02 15:04:05",
				"",
			} {
				timestamp := time.Date(2026, 11, 19, 12, 34, 56, nanos, zone)
				fm := NewTimeFormatter(layout)
				want := timestamp.Format(layout)
				if got := fm.FormatDateTime(timestamp); got != want {
					t.Fatalf("layout %q: got %q, want %q", layout, got, want)
				}
				buffer := fm.AppendDateTime([]byte("prefix:"), timestamp)
				if !bytes.Equal(buffer, []byte("prefix:"+want)) {
					t.Fatalf("AppendDateTime(%q) = %q", layout, buffer)
				}
			}
		}
	}
}

func TestSetTimeDecimalFormatPreservesTimezone(t *testing.T) {
	timestamp := time.Date(2026, 11, 19, 12, 34, 56, 7000000, time.UTC)
	fm := NewTimeFormatter(time.RFC3339Nano)
	fm.SetTimeDecimalFormat(3)
	if got, want := fm.FormatDateTime(timestamp), "2026-11-19T12:34:56.007Z"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	fm.SetTimeDecimalFormat(0)
	if got, want := fm.FormatDateTime(timestamp), "2026-11-19T12:34:56Z"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	fm.SetTimeDecimalFormat(6)
	if got, want := fm.FormatDateTime(timestamp), "2026-11-19T12:34:56.007000Z"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
