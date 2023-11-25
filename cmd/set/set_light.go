package set

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
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

# Turn off a light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --off

# Set brightness of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on --brightness 42.65
`
)

type LightOptions struct {
	Status     LightStatus
	Brightness float32
}

func NewSetLightOptions() *LightOptions {
	return &LightOptions{
		Status:     Undefined,
		Brightness: -1,
	}
}

// NewCmdSetLight returns initialized Command instance for the 'set light' sub command
func NewCmdSetLight(api *openhue.ClientWithResponses) *cobra.Command {

	o := NewSetLightOptions()

	cmd := &cobra.Command{
		Use:     "light [lightId]",
		Short:   docShort,
		Long:    docLong,
		Example: docExample,
		Args:    cobra.MatchAll(cobra.RangeArgs(1, 10), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(o.PrepareCmdSetLight(cmd))
			cobra.CheckErr(o.RunCmdSetLight(api, args))
		},
	}

	// on/off flags
	cmd.Flags().Bool("on", false, "Turn on the lights")
	cmd.Flags().Bool("off", false, "Turn off the lights")
	cmd.MarkFlagsMutuallyExclusive("on", "off")

	// Brightness flag
	cmd.Flags().Float32VarP(&o.Brightness, "brightness", "b", -1, "Set the brightness [min=0, max=100]")

	return cmd
}

// PrepareCmdSetLight makes sure provided values for LightOptions are valid
func (o *LightOptions) PrepareCmdSetLight(cmd *cobra.Command) error {

	on, _ := cmd.Flags().GetBool("on")
	off, _ := cmd.Flags().GetBool("off")

	if on {
		o.Status = On
	}

	if off {
		o.Status = Off
	}

	// validate the --brightness flag
	if o.Brightness > 100.0 || (o.Brightness != -1 && o.Brightness < 0) {
		return fmt.Errorf("--brightness flag must be greater than 0.0 and lower than 100.0, current value is %.2f", o.Brightness)
	}

	return nil
}

// RunCmdSetLight executes the light update command logic
func (o *LightOptions) RunCmdSetLight(api *openhue.ClientWithResponses, args []string) error {

	request := &openhue.UpdateLightJSONRequestBody{}

	if o.Status != Undefined {
		request.On = &openhue.On{
			On: ToBool(o.Status),
		}
	}

	if o.Brightness >= 0 && o.Brightness <= 100.0 {
		request.Dimming = &openhue.Dimming{
			Brightness: &o.Brightness,
		}
	}

	for _, id := range args {
		_, err := api.UpdateLight(context.Background(), id, *request)
		cobra.CheckErr(err)
	}

	return nil
}
