package cmd

import (
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
)

var (
	On  bool
	Off bool
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

		OnOff := true

		if On {
			OnOff = true
		}

		if Off {
			OnOff = false
		}

		request.On = &openhue.On{
			On: &OnOff,
		}

		for _, id := range args {
			_, err := openhue.Api.UpdateLight(context.Background(), id, *request)
			cobra.CheckErr(err)
		}
	},
}

func init() {
	setCmd.AddCommand(setLightCmd)

	// local flags
	setLightCmd.Flags().BoolVar(&On, "on", false, "Turn on the lights")
	setLightCmd.Flags().BoolVar(&Off, "off", false, "Turn off the lights")
	setLightCmd.MarkFlagsMutuallyExclusive("on", "off")
}
