package version

import (
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

			ctx.Io.Println("\n#  Version\t", ctx.BuildInfo.Version)
			ctx.Io.Println("#   Commit\t", BaseCommitUrl+ctx.BuildInfo.Commit)
			ctx.Io.Println("# Built at\t", ctx.BuildInfo.Date)
		},
	}

	return cmd
}
