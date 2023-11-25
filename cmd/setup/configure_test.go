package setup

import "testing"

func TestNewCmdConfigure(t *testing.T) {

	cmd := NewCmdConfigure()

	if cmd.Use != "configure" {
		t.Fatalf("The configure 'command' name has changed to '%s'", cmd.Use)
	}
}
