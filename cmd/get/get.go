package get

import (
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
	"openhue-cli/util"
)

type GetFlags struct {
	Json bool
}

var GetConfig GetFlags

// NewCmdGet returns an initialized Command instance for 'get' sub command
func NewCmdGet(api *openhue.ClientWithResponses) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Display one or many resources",
		GroupID: "hue",
		Long: `
Retrieve information for any kind of resources exposed by your Hue Bridge: lights, rooms, scenes, etc.
`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := api.GetResourcesWithResponse(context.Background())
			cobra.CheckErr(err)
			resources := *(*resp.JSON200).Data

			typeFlag := cmd.Flag("type").Value.String()

			if len(typeFlag) > 0 {
				// filter resources by type
				n := 0
				for _, r := range resources {
					if *r.Type == openhue.ResourceGetType(typeFlag) {
						resources[n] = r
						n++
					}
				}
				resources = resources[:n]
			}

			if GetConfig.Json {
				util.PrintJson(resources)
			} else {
				util.PrintTable(resources, PrintResource, "Resource ID", "Resource Type")
			}
		},
	}

	// local flags
	cmd.Flags().StringP("type", "t", "", "Filter by resource type (light, scene, room...)")

	// persistence flags
	cmd.PersistentFlags().BoolVar(&GetConfig.Json, "json", false, "Format output as JSON")

	// sub commands
	cmd.AddCommand(NewCmdGetLight(api))
	cmd.AddCommand(NewCmdGetRoom(api))

	return cmd
}

func PrintResource(resource openhue.ResourceGet) string {
	return *resource.Id + "\t" + string(*resource.Type)
}
