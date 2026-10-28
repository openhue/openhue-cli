package create

import (
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

// NewCmdCreate returns an initialized Command instance for the 'create' sub command
func NewCmdCreate(ctx *openhue.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create resources",
		GroupID: "hue",
		Long: `
Create new resources on your Hue Bridge.
`,
	}

	cmd.AddCommand(NewCmdCreateScene(ctx))

	return cmd
}
