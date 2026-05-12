package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"openhue-cli/openhue"
)

func TestRootIncludesCompletionSetupCommand(t *testing.T) {
	// Arrange
	t.Setenv("SHELL", "/bin/zsh")
	root, out := createTestContext()
	root.SetArgs([]string{"completion", "setup"})

	// Act
	err := root.Execute()

	// Assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "source <(openhue completion zsh)")
}

func createTestContext() (*cobra.Command, *bytes.Buffer) {
	ctx, out := openhue.NewTestContext(nil)
	root := NewCmdOpenHue(ctx)
	addCommands(root, ctx)

	return root, out
}
