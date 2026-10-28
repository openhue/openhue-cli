package update

import (
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

// NewCmdUpdate returns an initialized Command instance for the 'update' sub command
func NewCmdUpdate(ctx *openhue.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"u"},
		Short:   "Update resources",
		GroupID: "hue",
		Long: `
Update existing resources of your Hue Bridge.
`,
	}

	cmd.AddCommand(NewCmdUpdateScene(ctx))

	return cmd
}
