package util

import "time"

// NewTimer creates a new started Timer
func NewTimer() *Timer {
	return new(Timer).start()
}

// Timer can be used to monitor the execution time of a given method
type Timer struct {
	t0 time.Time
}

func (t *Timer) start() *Timer {
	t.t0 = time.Now()
	return t
}

// Reset sets the Timer initial time to time.Now()
func (t *Timer) Reset() *Timer {
	t.t0 = time.Now()
	return t
}

// SinceInMillis returns the elapsed time in milliseconds since the Timer was started or Reset
func (t *Timer) SinceInMillis() int64 {
	return time.Since(t.t0).Milliseconds()
}
