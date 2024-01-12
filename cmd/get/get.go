package get

import (
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
	"openhue-cli/openhue/gen"
	"openhue-cli/util"
)

type CmdGetOptions struct {
	Json bool
	Name bool
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
	cmd.PersistentFlags().BoolVarP(&o.Name, "name", "n", false, "Get resource(s) by name")

	// sub commands
	cmd.AddCommand(NewCmdGetLight(ctx, &o))
	cmd.AddCommand(NewCmdGetRoom(ctx, &o))

	return cmd
}

func RunGetAllResources(ctx *openhue.Context, typeFlag string, o *CmdGetOptions) {
	resp, err := ctx.Api.GetResourcesWithResponse(context.Background())
	cobra.CheckErr(err)
	resources := *(*resp.JSON200).Data

	if len(typeFlag) > 0 {
		// filter resources by type
		n := 0
		for _, r := range resources {
			if *r.Type == gen.ResourceGetType(typeFlag) {
				resources[n] = r
				n++
			}
		}
		resources = resources[:n]
	}

	if o.Json {
		util.PrintJson(ctx.Io, resources)
	} else {
		util.PrintTable(ctx.Io, resources, PrintResource, "Resource ID", "Resource Type")
	}
}

func PrintResource(resource gen.ResourceGet) string {
	return *resource.Id + "\t" + string(*resource.Type)
}
