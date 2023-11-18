package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
	"os"
)

// lightCmd represents the light command
var lightCmd = &cobra.Command{
	Use:   "light",
	Short: "Get light",
	Long: `
Fetches and displays all available lights
`,
	Args: cobra.MatchAll(cobra.RangeArgs(0, 1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {

		var lights *[]openhue.LightGet

		if len(args) > 0 {
			resp, err := openhue.Api.GetLightWithResponse(context.Background(), args[0])
			cobra.CheckErr(err)

			if resp.JSON200 == nil {
				fmt.Println("\nNot light found with ID", args[0])
				os.Exit(0)
			}

			lights = (*resp.JSON200).Data
		} else {
			resp, err := openhue.Api.GetLightsWithResponse(context.Background())
			cobra.CheckErr(err)
			lights = (*resp.JSON200).Data
		}

		if !GetConfig.Json {
			openhue.PrintTable(*lights, PrintLight, "Light ID", "Light Name", "Light Type")
		} else {
			openhue.PrintJsonArray(*lights)
		}
	},
}

func init() {
	getCmd.AddCommand(lightCmd)
}

func PrintLight(light openhue.LightGet) string {
	return *light.Id + "\t" + *light.Metadata.Name + "\t" + string(*light.Metadata.Archetype)
}
