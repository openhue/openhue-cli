package setup

import (
	"openhue-cli/openhue"
	"openhue-cli/openhue/test/assert"
	"testing"
)

func TestNewCmdDiscover(t *testing.T) {

	cmd := NewCmdDiscover(openhue.NewTestIOStreamsDiscard())

	assert.ThatCmdUseIs(t, cmd, "discover")
	assert.ThatCmdGroupIs(t, cmd, "config")
}
