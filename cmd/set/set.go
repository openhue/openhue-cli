package set

import (
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

// NewCmdSet returns an initialized Command instance for 'set' sub command
func NewCmdSet(api *openhue.ClientWithResponses) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Set specific features on resources",
		GroupID: "hue",
		Long: `
Set the values for a specific resource
`,
	}

	cmd.AddCommand(NewCmdSetLight(api))

	return cmd
}
