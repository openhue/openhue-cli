package openhue

import (
	"openhue-cli/openhue/test/assert"
	"testing"
)

func TestNewTestBuildInfo(t *testing.T) {

	info := NewTestBuildInfo()

	assert.Equals(t, "1.0.0", info.Version)
	assert.Equals(t, "1234", info.Commit)
	assert.Equals(t, "now", info.Date)
}
