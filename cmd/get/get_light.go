package get

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
	"openhue-cli/util"
	"os"
)

const (
	docLongGetLight = `
Fetches and displays all available lights
`
	docExampleGetLight = `
# List all lights as a table
openhue get light

# List all lights as JSON
openhue get light --json

# Get details for a single light
openhue get light aa31ba26-98a7-4830-8ae9-1b7caa8b5700 --json
`
)

type LightOptions struct {
	LightId string
}

func NewGetLightOptions() *LightOptions {
	return &LightOptions{}
}

// NewCmdGetLight returns initialized Command instance for the 'get light' sub command
func NewCmdGetLight(api *openhue.ClientWithResponses) *cobra.Command {

	o := NewGetLightOptions()

	cmd := &cobra.Command{
		Use:     "light [lightId]",
		Short:   "Get light",
		Long:    docLongGetLight,
		Example: docExampleGetLight,
		Args:    cobra.MatchAll(cobra.RangeArgs(0, 1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			o.PrepareGetRoomCmd(args)
			o.RunGetLightCmd(api)
		},
	}

	return cmd
}

func (o *LightOptions) PrepareGetRoomCmd(args []string) {
	if len(args) > 0 {
		o.LightId = args[0]
	}
}

func (o *LightOptions) RunGetLightCmd(api *openhue.ClientWithResponses) {
	var lights *[]openhue.LightGet

	if len(o.LightId) > 0 {
		resp, err := api.GetLightWithResponse(context.Background(), o.LightId)
		cobra.CheckErr(err)

		if resp.JSON200 == nil {
			fmt.Println("\nNot light found with ID", o.LightId)
			os.Exit(0)
		}

		lights = (*resp.JSON200).Data
	} else {
		resp, err := api.GetLightsWithResponse(context.Background())
		cobra.CheckErr(err)
		lights = (*resp.JSON200).Data
	}

	if !GetConfig.Json {
		util.PrintTable(*lights, PrintLight, "ID", "Name", "Type", "Status", "Brightness")
	} else {
		util.PrintJsonArray(*lights)
	}
}

func PrintLight(light openhue.LightGet) string {

	status := "[  ]"
	brightness := "N/A"

	if *light.On.On {
		status = "[on]"
	}

	if light.Dimming != nil {
		brightness = fmt.Sprint(*light.Dimming.Brightness) + "%"
	}

	return *light.Id + "\t" + *light.Metadata.Name + "\t" + string(*light.Metadata.Archetype) + "\t" + status + "\t" + brightness
}
