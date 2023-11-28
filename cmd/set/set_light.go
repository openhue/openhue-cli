package set

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
	"openhue-cli/openhue/gen"
	"openhue-cli/util/color"
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

# Set color (in RGB) of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on --rgb #3399FF

# Set color (in CIE space) of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on -x 0.675 -y 0.322

# Set color (by name) of a single light
openhue set light 15f51223-1e83-4e48-9158-0c20dbd5734e --on --color powder_blue
`
)

type LightOptions struct {
	Status     LightStatus
	Brightness float32
	Color      color.XY
}

type LightFlags struct {
	On         bool
	Off        bool
	Brightness float32
	Rgb        string
	X          float32
	Y          float32
	ColorName  string
}

func NewSetLightOptions() *LightOptions {
	return &LightOptions{
		Status:     Undefined,
		Brightness: -1,
		Color:      color.UndefinedColor,
	}
}

// NewCmdSetLight returns initialized Command instance for the 'set light' sub command
func NewCmdSetLight(ctx *openhue.Context) *cobra.Command {

	o := NewSetLightOptions()
	f := LightFlags{}

	cmd := &cobra.Command{
		Use:     "light [lightId]",
		Short:   docShort,
		Long:    docLong,
		Example: docExample,
		Args:    cobra.MatchAll(cobra.RangeArgs(1, 10), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(o.PrepareCmdSetLight(&f))
			cobra.CheckErr(o.RunCmdSetLight(ctx.Api, args))
		},
	}

	f.InitFlags(cmd)

	return cmd
}

func (f *LightFlags) InitFlags(cmd *cobra.Command) {
	// on/off flags
	cmd.Flags().BoolVar(&f.On, "on", false, "Turn on the lights")
	cmd.Flags().BoolVar(&f.Off, "off", false, "Turn off the lights")
	cmd.MarkFlagsMutuallyExclusive("on", "off")

	// Brightness flag
	cmd.Flags().Float32VarP(&f.Brightness, "brightness", "b", -1, "Set the brightness [min=0, max=100]")

	//
	// Color flags
	//

	// rgb
	cmd.Flags().StringVar(&f.Rgb, "rgb", "", "RGB hexadecimal value (example #CCE5FF)")

	// xy
	cmd.Flags().Float32VarP(&f.X, "cie-x", "x", -1, "X coordinate in the CIE color space (example 0.675)")
	cmd.Flags().Float32VarP(&f.Y, "cie-y", "y", -1, "Y coordinate in the CIE color space (example 0.322)")
	cmd.MarkFlagsRequiredTogether("cie-x", "cie-y")

	// name
	cmd.Flags().StringVarP(&f.ColorName, "color", "c", "", "Color name. Allowed: white, lime, green, blue, etc.")
	_ = cmd.RegisterFlagCompletionFunc("color", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return color.GetSupportColorList(), cobra.ShellCompDirectiveDefault
	})

	// exclusions
	cmd.MarkFlagsMutuallyExclusive("color", "rgb", "cie-x")
}

// PrepareCmdSetLight makes sure provided values for LightOptions are valid
func (o *LightOptions) PrepareCmdSetLight(flags *LightFlags) error {

	if flags.On {
		o.Status = On
	}

	if flags.Off {
		o.Status = Off
	}

	// validate the --brightness flag
	if flags.Brightness > 100.0 || (flags.Brightness != -1 && flags.Brightness < 0) {
		return fmt.Errorf("--brightness flag must be greater than 0.0 and lower than 100.0, current value is %.2f", o.Brightness)
	} else {
		o.Brightness = flags.Brightness
	}

	// color in RGB
	if flags.Rgb != "" {
		rgb, err := color.NewRGBFomHex(flags.Rgb)
		if err != nil {
			return err
		}
		o.Color = *rgb.ToXY()
	}

	// color in CIE space
	if flags.X >= 0 && flags.Y >= 0 {
		o.Color = color.XY{
			X: flags.X,
			Y: flags.Y,
		}
	}

	// color from enum
	if flags.ColorName != "" {
		c, err := color.FindColorByName(flags.ColorName)
		cobra.CheckErr(err)
		o.Color = color.XY{
			X: c.X,
			Y: c.Y,
		}
	}

	return nil
}

// RunCmdSetLight executes the light update command logic
func (o *LightOptions) RunCmdSetLight(api *gen.ClientWithResponses, args []string) error {

	request := &gen.UpdateLightJSONRequestBody{}

	if o.Status != Undefined {
		request.On = &gen.On{
			On: ToBool(o.Status),
		}
	}

	if o.Brightness >= 0 && o.Brightness <= 100.0 {
		request.Dimming = &gen.Dimming{
			Brightness: &o.Brightness,
		}
	}

	if o.Color != color.UndefinedColor {
		request.Color = &gen.Color{
			Xy: &gen.GamutPosition{
				X: &o.Color.X,
				Y: &o.Color.Y,
			},
		}
	}

	for _, lightId := range args {
		_, err := api.UpdateLight(context.Background(), lightId, *request)
		cobra.CheckErr(err)
	}

	return nil
}
