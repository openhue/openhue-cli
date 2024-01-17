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
openhue get light

# List all lights as JSON
openhue get light --json

# Get details for a single light
openhue get light aa31ba26-98a7-4830-8ae9-1b7caa8b5700 --json

# Get light by name
openhue get light -n "Hue Go"
`
)

type LightOptions struct {
	LightParam string // ID or Name
	Json       *bool
	Name       *bool
}

func NewGetLightOptions(co *CmdGetOptions) *LightOptions {
	return &LightOptions{
		Json: &co.Json,
		Name: &co.Name,
	}
}

// NewCmdGetLight returns initialized Command instance for the 'get light' sub command
func NewCmdGetLight(ctx *openhue.Context, co *CmdGetOptions) *cobra.Command {

	o := NewGetLightOptions(co)

	cmd := &cobra.Command{
		Use:     "light [lightId]",
		Aliases: []string{"lights"},
		Short:   "Get light",
		Long:    docLongGetLight,
		Example: docExampleGetLight,
		Args:    cobra.MatchAll(cobra.RangeArgs(0, 1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			o.PrepareGetLightCmd(args)
			o.RunGetLightCmd(ctx)
		},
	}

	return cmd
}

func (o *LightOptions) PrepareGetLightCmd(args []string) {
	if len(args) > 0 {
		o.LightParam = args[0]
	}
}

func (o *LightOptions) RunGetLightCmd(ctx *openhue.Context) {
	var lights []openhue.Light

	if len(o.LightParam) > 0 {

		if *o.Name {
			lights = openhue.FindLightsByName(ctx.Home, []string{o.LightParam})
		} else {
			lights = openhue.FindLightsByIds(ctx.Home, []string{o.LightParam})
		}

	} else {
		lights = openhue.FindAllLights(ctx.Home)
	}

	if *o.Json {
		util.PrintJsonArray(ctx.Io, lights)
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

	return light.Id + "\t" + light.Name + "\t" + string(*light.HueData.Metadata.Archetype) + "\t" + status + "\t" + brightness + "\t" + room
}
