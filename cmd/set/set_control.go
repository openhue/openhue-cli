package set

import (
	"context"
	"fmt"
	"openhue-cli/openhue"
	"strconv"
	"time"

	oh "github.com/openhue/openhue-go"
	"github.com/spf13/cobra"
)

const (
	setControlDocShort = "Configure a control (button)"
	setControlDocLong  = `
Change the settings of a control, as listed by 'openhue get control'.

The control is selected either by its ID, or by the name or ID of its device together with --control
(the control number on the device). Only buttons have settings today: rotary dials are not configurable.
`
	setControlDocExample = `
# Set the repeat interval of a button, in milliseconds
openhue set control 462e54d9-ec5d-4bf6-879d-ad34cb9a692e --repeat-interval 800

# Same, with a duration
openhue set control 462e54d9-ec5d-4bf6-879d-ad34cb9a692e --repeat-interval 800ms

# Select the button by device name and control number
openhue set control "Hue dimmer switch" --control 2 --repeat-interval 1s`
)

// CmdSetControlOptions holds the settings 'set control' can change: add new fields here as the bridge exposes them.
type CmdSetControlOptions struct {
	control        int
	repeatInterval string
}

// NewCmdSetControl returns initialized cobra.Command instance for the 'set control' sub command
func NewCmdSetControl(ctx *openhue.Context) *cobra.Command {

	o := &CmdSetControlOptions{}

	cmd := &cobra.Command{
		Use:     "control [controlId|deviceName]",
		Aliases: []string{"controls"},
		Short:   setControlDocShort,
		Long:    setControlDocLong,
		Example: setControlDocExample,
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(ctx, cmd, args[0])
		},
	}

	cmd.Flags().IntVar(&o.control, "control", 0, "control number on the device, when selecting the control by device")
	cmd.Flags().StringVar(&o.repeatInterval, "repeat-interval", "", "button repeat interval while held: a duration (500ms, 1s) or a number of milliseconds")

	return cmd
}

func (o *CmdSetControlOptions) run(ctx *openhue.Context, cmd *cobra.Command, target string) error {

	// validate the flags before resolving the control, nothing is sent on failure
	if !cmd.Flags().Changed("repeat-interval") {
		return fmt.Errorf("no setting to change, available flags: --repeat-interval")
	}
	interval, err := parseRepeatInterval(o.repeatInterval)
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("control") && o.control <= 0 {
		return fmt.Errorf("--control must be a positive control number, got %d", o.control)
	}

	control, err := openhue.FindControl(ctx.Home, target, o.control)
	if err != nil {
		return err
	}

	if control.Kind != openhue.ControlKindButton {
		return fmt.Errorf("control %s is a %s: rotary dials have no configurable settings", control.Id, control.Kind)
	}

	body := oh.ButtonPut{
		Button: &struct {
			RepeatInterval *int `json:"repeat_interval,omitempty"`
		}{RepeatInterval: &interval},
	}
	if err := ctx.H.UpdateButton(context.Background(), control.Id, body); err != nil {
		cmd.SilenceUsage = true
		return err
	}

	ctx.Io.Printf("Control '%s' (%s) repeat interval set to %dms\n", control.Name, control.Id, interval)
	return nil
}

// parseRepeatInterval accepts a Go duration ("500ms", "1s") or a plain number of milliseconds, and returns
// a strictly positive number of milliseconds.
func parseRepeatInterval(value string) (int, error) {
	var d time.Duration
	if ms, err := strconv.Atoi(value); err == nil {
		d = time.Duration(ms) * time.Millisecond
	} else if d, err = time.ParseDuration(value); err != nil {
		return 0, fmt.Errorf("invalid --repeat-interval '%s': use a duration (500ms, 1s) or a number of milliseconds", value)
	}

	ms := int(d.Milliseconds())
	if ms <= 0 {
		return 0, fmt.Errorf("invalid --repeat-interval '%s': must be greater than 0 ms", value)
	}
	return ms, nil
}
