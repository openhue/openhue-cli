package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
)

var (
	OnOff bool
)

// setLightCmd represents the setLight command
var setLightCmd = &cobra.Command{
	Use:   "light",
	Short: "Update one or multiple lights",
	Long: `
Update one or multiple lights (max is 10 lights simultaneously). Allows to turn on or off a light.
`,
	Args: cobra.MatchAll(cobra.RangeArgs(1, 10), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {

		request := &openhue.UpdateLightJSONRequestBody{}

		request.On = &openhue.On{
			On: &OnOff,
		}

		for _, id := range args {
			_, err := openhue.Api.UpdateLight(context.Background(), id, *request)
			cobra.CheckErr(err)
			fmt.Println("Light", id, "successfully updated")
		}

	},
}

func init() {
	setCmd.AddCommand(setLightCmd)

	// local flags
	setLightCmd.Flags().BoolVar(&OnOff, "on", true, "Turn on or off the lights")
}
