package get

import (
	oh "github.com/openhue/openhue-go"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
	"openhue-cli/util"
)

type CmdGetOptions struct {
	Json bool
}

// NewCmdGet returns an initialized Command instance for 'get' sub command
func NewCmdGet(ctx *openhue.Context) *cobra.Command {

	o := CmdGetOptions{}

	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "Display one or many resources",
		GroupID: "hue",
		Long: `
Retrieve information for any kind of resources exposed by your Hue Bridge: lights, rooms, scenes, etc.
`,
		Run: func(cmd *cobra.Command, args []string) {
			RunGetAllResources(ctx, cmd.Flag("type").Value.String(), &o)
		},
	}

	// local flags
	cmd.Flags().StringP("type", "t", "", "Filter by resource type (light, scene, room...)")

	// persistence flags
	cmd.PersistentFlags().BoolVarP(&o.Json, "json", "j", false, "Format output as JSON")

	// sub commands
	cmd.AddCommand(NewCmdGetEvents(ctx, &o))
	cmd.AddCommand(NewCmdGetLight(ctx, &o))
	cmd.AddCommand(NewCmdGetRoom(ctx, &o))
	cmd.AddCommand(NewCmdGetScene(ctx, &o))

	return cmd
}

func RunGetAllResources(ctx *openhue.Context, typeFlag string, o *CmdGetOptions) {

	resp, err := ctx.H.GetResources()
	cobra.CheckErr(err)
	var resources []oh.ResourceGet
	for _, r := range resp {
		if len(typeFlag) == 0 || len(typeFlag) > 0 && *r.Type == oh.ResourceGetType(typeFlag) {
			resources = append(resources, r)
		}
	}

	if o.Json {
		util.PrintJson(ctx.Io, resources)
	} else {
		util.PrintTable(ctx.Io, resources, PrintResource, "Resource ID", "Resource Type")
	}
}

func PrintResource(resource oh.ResourceGet) string {
	return *resource.Id + "\t" + string(*resource.Type)
}
