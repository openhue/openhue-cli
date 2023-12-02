package setup

import (
	"openhue-cli/openhue"
	"openhue-cli/openhue/test/assert"
	"testing"
)

func TestNewCmdAuth(t *testing.T) {

	cmd := NewCmdAuth(openhue.NewTestIOStreamsDiscard())

	assert.ThatCmdUseIs(t, cmd, "auth")
	assert.ThatCmdGroupIs(t, cmd, "config")
}
