package get

import (
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
	"openhue-cli/openhue/gen"
	"openhue-cli/util"
)

type GetFlags struct {
	Json bool
}

var GetConfig GetFlags

// NewCmdGet returns an initialized Command instance for 'get' sub command
func NewCmdGet(ctx *openhue.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Display one or many resources",
		GroupID: "hue",
		Long: `
Retrieve information for any kind of resources exposed by your Hue Bridge: lights, rooms, scenes, etc.
`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := ctx.Api.GetResourcesWithResponse(context.Background())
			cobra.CheckErr(err)
			resources := *(*resp.JSON200).Data

			typeFlag := cmd.Flag("type").Value.String()

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

			if GetConfig.Json {
				util.PrintJson(ctx.Io, resources)
			} else {
				util.PrintTable(ctx.Io, resources, PrintResource, "Resource ID", "Resource Type")
			}
		},
	}

	// local flags
	cmd.Flags().StringP("type", "t", "", "Filter by resource type (light, scene, room...)")

	// persistence flags
	cmd.PersistentFlags().BoolVar(&GetConfig.Json, "json", false, "Format output as JSON")

	// sub commands
	cmd.AddCommand(NewCmdGetLight(ctx))
	cmd.AddCommand(NewCmdGetRoom(ctx))

	return cmd
}

func PrintResource(resource gen.ResourceGet) string {
	return *resource.Id + "\t" + string(*resource.Type)
}
