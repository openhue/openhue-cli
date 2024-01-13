package setup

import (
	"openhue-cli/openhue"
	"openhue-cli/openhue/test/assert"
	"testing"
)

func TestNewCmdSetup(t *testing.T) {

	cmd := NewCmdSetup(openhue.NewTestIOStreamsDiscard())

	assert.ThatCmdUseIs(t, cmd, "setup")
	assert.ThatCmdGroupIs(t, cmd, "config")

}
