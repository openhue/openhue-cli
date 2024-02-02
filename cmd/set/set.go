package set

import (
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

// NewCmdSet returns an initialized Command instance for 'set' sub command
func NewCmdSet(ctx *openhue.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "set",
		Aliases: []string{"s"},
		Short:   "Set specific features on resources",
		GroupID: "hue",
		Long: `
Set the values for specific resources.
`,
	}

	cmd.AddCommand(NewCmdSetLight(ctx))
	cmd.AddCommand(NewCmdSetRoom(ctx))
	cmd.AddCommand(NewCmdSetScene(ctx))

	return cmd
}
