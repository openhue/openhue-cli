package version

import (
	"fmt"
	"github.com/spf13/cobra"
	"openhue-cli/util"
)

func NewCmdVersion(buildInfo *util.BuildInfo) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version information`,
		Example: `
# Print the version
openhue version
`,
		Run: func(cmd *cobra.Command, args []string) {

			fmt.Println("\n# Version\t", buildInfo.Version)
			fmt.Println("#  Commit\t", buildInfo.Commit)
			fmt.Println("#    Time\t", buildInfo.Date)
		},
	}

	return cmd

}
