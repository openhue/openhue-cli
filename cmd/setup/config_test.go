package setup

import (
	"openhue-cli/openhue/test/assert"
	"testing"
)

func TestNewCmdConfigure(t *testing.T) {

	cmd := NewCmdConfigure()

	assert.ThatCmdUseIs(t, cmd, "config")
	assert.ThatCmdGroupIs(t, cmd, "config")

}
