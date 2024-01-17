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

# Turn on multiple lights
openhue set light 83111103-a3eb-40c5-b22a-02deedd21fcb 8f0a7b52-df25-4bc7-b94d-0dd1a88068ff --on

# Turn off a light identified by its name
openhue set light -n "Hue Play TV" --off

# Set brightness of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on --brightness 42.65

# Set color (in RGB) of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on --rgb #3399FF

# Set color (in CIE space) of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on -x 0.675 -y 0.322

# Set color (by name) of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on --color powder_blue
`
)

// NewCmdSetLight returns initialized Command instance for the 'set light' sub command
func NewCmdSetLight(ctx *openhue.Context, setOpt *CmdSetOptions) *cobra.Command {

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

			var lights []openhue.Light

			if setOpt.Name {
				lights = openhue.FindLightsByName(ctx.Home, args)
			} else {
				lights = openhue.FindLightsByIds(ctx.Home, args)
			}

			if len(lights) == 0 {
				ctx.Io.ErrPrintln("no light(s) found for given ID(s)", args)
			}

			for _, light := range lights {
				log.Info("set light ", light.Id)
				light.Set(o)
			}
		},
	}

	f.initCmd(cmd)

	return cmd
}
