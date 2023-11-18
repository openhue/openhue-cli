package cmd

import (
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
)

type GetFlags struct {
	Json bool
}

var GetConfig GetFlags

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources",
	Long: `
Retrieve information for any kind of resources exposed by your Hue Bridge: lights, rooms, scenes, etc.
`,
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := openhue.Api.GetResourcesWithResponse(context.Background())
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
			openhue.PrintJson(resources)
		} else {
			openhue.PrintTable(resources, PrintResource, "Resource ID", "Resource Type")
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// local flags
	getCmd.Flags().StringP("type", "t", "", "Filter by resource type (light, scene, room...)")

	// persistence flags
	getCmd.PersistentFlags().BoolVar(&GetConfig.Json, "json", false, "Format output as JSON")
}

func PrintResource(resource openhue.ResourceGet) string {
	return *resource.Id + "\t" + string(*resource.Type)
}
