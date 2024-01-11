package assert

import (
	"github.com/spf13/cobra"
	"testing"
)

// ThatLineEqualsTo verifies that an `idx` line contained in the `lines` slice should be equal to the `expected` value
func ThatLineEqualsTo(t *testing.T, lines []string, idx int, expected string) {
	if lines[idx] != expected {
		t.Fatalf("expected \"%s\", obtained \"%s\"", expected, lines[idx])
	}
}

// ThatCmdUseIs verifies that the command Use is valid
func ThatCmdUseIs(t *testing.T, cmd *cobra.Command, expectedUse string) {
	if cmd.Use != expectedUse {
		t.Fatalf("wrong command use. Expected '%s', got '%s'", expectedUse, cmd.Use)
	}
}

// ThatCmdGroupIs verifies that the command GroupID is valid
func ThatCmdGroupIs(t *testing.T, cmd *cobra.Command, expectedGroup string) {
	if cmd.GroupID != expectedGroup {
		t.Fatalf("wrong command group. Expected '%s', got '%s'", expectedGroup, cmd.Use)
	}
}
