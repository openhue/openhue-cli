package openhue

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTestBuildInfo(t *testing.T) {

	info := NewTestBuildInfo()

	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "1234", info.Commit)
	assert.Equal(t, "now", info.Date)
}
