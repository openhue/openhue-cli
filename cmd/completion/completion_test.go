package completion

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionCommandsGenerateScripts(t *testing.T) {
	tests := []struct {
		name          string
		shell         string
		expectedLines []string
	}{
		{
			name:  "bash",
			shell: "bash",
			expectedLines: []string{
				"# bash completion V2 for openhue",
				"__start_openhue",
			},
		},
		{
			name:  "zsh",
			shell: "zsh",
			expectedLines: []string{
				"#compdef openhue",
				"_openhue()",
			},
		},
		{
			name:  "fish",
			shell: "fish",
			expectedLines: []string{
				"# fish completion for openhue",
				"complete -c openhue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			root, out := createTestContext("/bin/zsh")
			root.SetArgs([]string{"completion", tt.shell})

			// Act
			err := root.Execute()

			// Assert
			require.NoError(t, err)
			for _, expectedLine := range tt.expectedLines {
				assert.Contains(t, out.String(), expectedLine)
			}
		})
	}
}

func TestCompletionSetupPrintsCommandForDetectedShell(t *testing.T) {
	tests := []struct {
		name            string
		detectedShell   string
		expectedCommand string
	}{
		{
			name:            "bash",
			detectedShell:   "/opt/homebrew/bin/bash",
			expectedCommand: "source <(openhue completion bash)",
		},
		{
			name:            "zsh",
			detectedShell:   "/bin/zsh",
			expectedCommand: "source <(openhue completion zsh)",
		},
		{
			name:            "fish",
			detectedShell:   "/opt/homebrew/bin/fish",
			expectedCommand: "openhue completion fish | source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			root, out := createTestContext(tt.detectedShell)
			root.SetArgs([]string{"completion", "setup"})

			// Act
			err := root.Execute()

			// Assert
			require.NoError(t, err)
			assert.Contains(t, out.String(), tt.expectedCommand)
		})
	}
}

func TestCompletionSetupReturnsAnErrorForUnsupportedShell(t *testing.T) {
	// Arrange
	root, out := createTestContext("/bin/tcsh")
	root.SetArgs([]string{"completion", "setup"})

	// Act
	err := root.Execute()

	// Assert
	require.Error(t, err)
	assert.Empty(t, strings.TrimSpace(out.String()))
	assert.Contains(t, err.Error(), "unsupported shell")
}

func createTestContext(shell string) (*cobra.Command, *strings.Builder) {
	out := &strings.Builder{}
	root := &cobra.Command{Use: "openhue"}
	root.SetOut(out)
	root.SetErr(out)
	root.AddGroup(&cobra.Group{
		ID:    "config",
		Title: "Configuration",
	})
	root.AddCommand(newCmdCompletion(out, func() string {
		return shell
	}))

	return root, out
}
