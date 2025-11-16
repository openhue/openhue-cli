package get

import (
	"fmt"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
	"openhue-cli/util"
)

const (
	docLongGetLight = `
Fetches and displays all available lights
`
	docExampleGetLight = `
# List all lights as a table
openhue get lights

# List all lights as JSON
openhue get lights --json

# Get details for a single light
openhue get light aa31ba26-98a7-4830-8ae9-1b7caa8b5700 --json

# Get light by name
openhue get light "Hue Go"

# List all lights for a specific room
openhue get light --room "Living Room"
`
)

type LightOptions struct {
	*CmdGetOptions
	Room string
}

// NewCmdGetLight returns initialized Command instance for the 'get light' sub command
func NewCmdGetLight(ctx *openhue.Context, co *CmdGetOptions) *cobra.Command {

	o := LightOptions{
		CmdGetOptions: co,
	}

	cmd := &cobra.Command{
		Use:     "light [lightId]",
		Aliases: []string{"lights"},
		Short:   "Get light",
		Long:    docLongGetLight,
		Example: docExampleGetLight,
		Args:    cobra.MatchAll(cobra.RangeArgs(0, 1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			o.RunGetLightCmd(ctx, args)
		},
	}

	cmd.Flags().StringVarP(&o.Room, "room", "r", "", "Filter scenes by room (name or ID)")

	return cmd
}

func (o *LightOptions) RunGetLightCmd(ctx *openhue.Context, args []string) {

	lights := openhue.SearchLights(ctx.Home, o.Room, args)

	if o.Json {
		util.PrintJsonArray(ctx.Io, lights)
	} else if o.Terminal {
		util.PrintTerminal(ctx.Io, lights, PrintLight)
	} else {
		util.PrintTable(ctx.Io, lights, PrintLight, "ID", "Name", "Type", "Status", "Brightness", "Room")
	}
}

func PrintLight(light openhue.Light) string {

	status := "[  ]"
	brightness := "N/A"
	room := light.Parent.Parent.Name // parent of a light is the device that belongs to a room

	if *light.HueData.On.On {
		status = "[on]"
	}

	if light.HueData.Dimming != nil {
		brightness = fmt.Sprint(*light.HueData.Dimming.Brightness) + "%"
	}

	return light.Id +
		"\t" + light.Name +
		"\t" + string(*light.HueData.Metadata.Archetype) +
		"\t" + status +
		"\t" + brightness +
		"\t" + room
}
