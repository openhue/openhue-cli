package version

import (
	"openhue-cli/openhue"
	"openhue-cli/openhue/test"
	"strings"
	"testing"
)

const (
	Line1 = "#  Version\t 1.0.0"
	Line2 = "#   Commit\t https://github.com/openhue/openhue-cli/commit/1234"
	Line3 = "# Built at\t now"
)

func TestNewCmdVersion(t *testing.T) {

	ctx, out := openhue.NewTestContextWithoutApi()

	cmd := NewCmdVersion(ctx)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute the `version` command: %s", err)
	}

	lines := strings.Split(out.String(), "\n")

	test.AssertThatLineEqualsTo(t, lines, 1, Line1)
	test.AssertThatLineEqualsTo(t, lines, 2, Line2)
	test.AssertThatLineEqualsTo(t, lines, 3, Line3)

}
