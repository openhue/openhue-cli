package util

import "time"

func NewTimer() *Timer {
	return new(Timer).start()
}

type Timer struct {
	t0 time.Time
}

func (t *Timer) start() *Timer {
	t.t0 = time.Now()
	return t
}

func (t *Timer) SinceInMillis() int64 {
	return time.Since(t.t0).Milliseconds()
}
