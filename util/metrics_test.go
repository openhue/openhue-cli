package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewTimer(t *testing.T) {
	timer := NewTimer()
	assert.NotNil(t, timer)
	time.Sleep(1 * time.Millisecond)
	assert.True(t, timer.SinceInMillis() >= 1, "we slept for 1 millis")
	timer.Reset()
	time.Sleep(2 * time.Millisecond)
	assert.True(t, timer.SinceInMillis() >= 2, "we slept for 2 millis")
}
