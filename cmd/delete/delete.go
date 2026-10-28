package delete

import (
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

// NewCmdDelete returns an initialized Command instance for the 'delete' sub command
func NewCmdDelete(ctx *openhue.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"d"},
		Short:   "Delete resources",
		GroupID: "hue",
		Long: `
Delete resources from your Hue Bridge.
`,
	}

	cmd.AddCommand(NewCmdDeleteScene(ctx))

	return cmd
}
