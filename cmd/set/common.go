package set

import (
	"fmt"
	"openhue-cli/openhue"
	"openhue-cli/util/color"
	"time"

	"github.com/spf13/cobra"
)

//
// Set Light flags, used by the 'set light' and 'set room' commands
//

type CmdSetLightFlags struct {
	On             bool
	Off            bool
	Brightness     float32
	Rgb            string
	X              float32
	Y              float32
	ColorName      string
	Temperature    int
	TransitionTime string
}

func (flags *CmdSetLightFlags) initCmd(cmd *cobra.Command) {
	// on/off flags
	cmd.Flags().BoolVar(&flags.On, "on", false, "Turn on the lights")
	cmd.Flags().BoolVar(&flags.Off, "off", false, "Turn off the lights")
	cmd.MarkFlagsMutuallyExclusive("on", "off")

	// Brightness flag
	cmd.Flags().Float32VarP(&flags.Brightness, "brightness", "b", -1, "Set the brightness [min=0, max=100]")

	//
	// Color flags
	//

	// rgb
	cmd.Flags().StringVar(&flags.Rgb, "rgb", "", "RGB hexadecimal value (example #CCE5FF)")

	// xy
	cmd.Flags().Float32VarP(&flags.X, "cie-x", "x", -1, "X coordinate in the CIE color space (example 0.675)")
	cmd.Flags().Float32VarP(&flags.Y, "cie-y", "y", -1, "Y coordinate in the CIE color space (example 0.322)")
	cmd.MarkFlagsRequiredTogether("cie-x", "cie-y")

	// name
	cmd.Flags().StringVarP(&flags.ColorName, "color", "c", "", "Color name. Allowed: white, lime, green, blue, etc.")
	_ = cmd.RegisterFlagCompletionFunc("color", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return color.GetSupportColorList(), cobra.ShellCompDirectiveDefault
	})

	// temperature
	cmd.Flags().IntVarP(&flags.Temperature, "temperature", "t", -1, "Color temperature in Mirek [min=153, max=500]")

	// transition time
	cmd.Flags().StringVar(&flags.TransitionTime, "transition-time", "0s", "Duration of a light transition")

	// exclusions
	cmd.MarkFlagsMutuallyExclusive("color", "rgb", "cie-x", "temperature")
}

// toSetLightOptions makes sure provided values for LightOptions are valid
func (flags *CmdSetLightFlags) toSetLightOptions() (*openhue.SetLightOptions, error) {

	o := openhue.NewSetLightOptions()

	if flags.On {
		o.Status = openhue.LightStatusOn
	}

	if flags.Off {
		o.Status = openhue.LightStatusOff
	}

	// validate the --brightness flag
	if flags.Brightness > 100.0 || (flags.Brightness != -1 && flags.Brightness < 0) {
		return nil, fmt.Errorf("--brightness flag must be greater than 0.0 and lower than 100.0, current value is %.2f", flags.Brightness)
	} else {
		o.Brightness = flags.Brightness
	}

	// validate the temperature flag
	if flags.Temperature > 500 || (flags.Temperature != -1 && flags.Temperature < 153) {
		return nil, fmt.Errorf("--temperature flag must be greater than 153 and lower than 500, current value is %d", flags.Temperature)
	} else {
		o.Temperature = flags.Temperature
	}

	// color in RGB
	if flags.Rgb != "" {
		rgb, err := color.NewRGBFomHex(flags.Rgb)
		if err != nil {
			return nil, err
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

	if flags.TransitionTime != "0s" {
		duration, err := time.ParseDuration(flags.TransitionTime)
		cobra.CheckErr(err)
		o.TransitionTime = int(duration.Milliseconds())
	}

	return o, nil
}
