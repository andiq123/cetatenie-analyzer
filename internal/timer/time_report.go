package timer

import (
	"fmt"
	"strings"
	"time"
)

type TimeReport struct {
	FetchTime time.Duration
	ParseTime time.Duration
}

func NewTimeReport(fetchTime, parseTime time.Duration) *TimeReport {
	return &TimeReport{FetchTime: fetchTime, ParseTime: parseTime}
}

func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0 secunde"
	}
	s := fmt.Sprintf("%.2f secunde", d.Seconds())
	return strings.NewReplacer(
		".", "\\.",
		"-", "\\-",
		"(", "\\(",
		")", "\\)",
		"!", "\\!",
	).Replace(s)
}

type Timer struct {
	start time.Time
	end   time.Time
}

func NewTimer() *Timer {
	return &Timer{}
}

func (t *Timer) Start() {
	t.start = time.Now()
}

func (t *Timer) Stop() {
	t.end = time.Now()
}

func (t *Timer) Duration() time.Duration {
	return t.end.Sub(t.start)
}
