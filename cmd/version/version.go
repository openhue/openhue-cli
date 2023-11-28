package version

import (
	"fmt"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

const (
	BaseCommitUrl = "https://github.com/openhue/openhue-cli/commit/"
)

func NewCmdVersion(ctx *openhue.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version information`,
		Example: `
# Print the version
openhue version
`,
		Run: func(cmd *cobra.Command, args []string) {

			fmt.Fprintln(ctx.Io.Out, "\n#  Version\t", ctx.BuildInfo.Version)
			fmt.Fprintln(ctx.Io.Out, "#   Commit\t", BaseCommitUrl+ctx.BuildInfo.Commit)
			fmt.Fprintln(ctx.Io.Out, "# Built at\t", ctx.BuildInfo.Date)
		},
	}

	return cmd
}
