package set

import (
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

// CmdSetOptions contains common flags for all 'set' sub commands
type CmdSetOptions struct {
	Name bool
}

// NewCmdSet returns an initialized Command instance for 'set' sub command
func NewCmdSet(ctx *openhue.Context) *cobra.Command {

	o := &CmdSetOptions{}

	cmd := &cobra.Command{
		Use:     "set",
		Aliases: []string{"s"},
		Short:   "Set specific features on resources",
		GroupID: "hue",
		Long: `
Set the values for a specific resource
`,
	}

	cmd.PersistentFlags().BoolVarP(&o.Name, "name", "n", false, "Set resource(s) by name")

	cmd.AddCommand(NewCmdSetLight(ctx, o))
	cmd.AddCommand(NewCmdSetRoom(ctx, o))

	return cmd
}
