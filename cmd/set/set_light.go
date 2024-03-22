package set

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

const (
	docShort = "Update one or multiple lights (on/off, brightness, color)"
	docLong  = `
Update one or multiple lights (max is 10 lights simultaneously).

 You must set the --on flag in order to turn a light on, even if you set the brightness or the color values.
    
Use "openhue get light" for a complete list of available lights.
`
	docExample = `
# Turn on a light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on

# Turn on multiple lights, using either their ID or their name
openhue set light 83111103-a3eb-40c5-b22a-02deedd21fcb "Hue Play TV" --on

# Turn off a light identified by its name in a given room. This is useful when there are multiple lights with the same name.
openhue set light --room "Living Room" "Hue Play Right" --off

# Set brightness of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on --brightness 42.65

# Set color (in RGB) of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on --rgb #3399FF

# Set color (in CIE space) of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on -x 0.675 -y 0.322

# Set color (by name) of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on --color powder_blue

# Set color temperature (in Mirek) of a single light
openhue set light MyLight --on -t 250
`
)

var room = ""

// NewCmdSetLight returns initialized Command instance for the 'set light' sub command
func NewCmdSetLight(ctx *openhue.Context) *cobra.Command {

	f := CmdSetLightFlags{}

	cmd := &cobra.Command{
		Use:     "light [lightId]",
		Aliases: []string{"lights"},
		Short:   docShort,
		Long:    docLong,
		Example: docExample,
		Args:    cobra.MatchAll(cobra.RangeArgs(1, 10), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			o, err := f.toSetLightOptions()
			cobra.CheckErr(err)

			lights := openhue.SearchLights(ctx.Home, room, args)

			if len(lights) == 0 {
				ctx.Io.ErrPrintln("no light(s) found for", args, "in room", room)
				return
			}

			for _, light := range lights {
				log.Info("set light ", light.Id)
				light.Set(o)
			}
		},
	}

	f.initCmd(cmd)

	cmd.Flags().StringVarP(&room, "room", "r", "", "Force the room of the light(s). Useful when there are multiple lights with the same name.")

	return cmd
}
