package set

import (
	"context"
	"fmt"
	"openhue-cli/openhue"
	"openhue-cli/util"
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

Besides the repeat interval, you can choose what the button does, as the Hue app does:
  --scene          a short press recalls this scene (name or ID, among the scenes of the room the button is
                   bound to). Repeat it to cycle through several scenes, in the given order.
  --with-off       a second press turns the lights off
  --repeat-timeout idle time after which a scene cycle restarts (default 3s)
  --on-hold        what holding the button does: dim-up, dim-down, dim-alternate or none
  --clear          remove the short-press action
  --dry-run        print the configuration that would be sent to the bridge, and send nothing

Everything the flags do not cover (other buttons, long press, time-based assignments...) is left unchanged.
The accessory must already be configured on the bridge, for instance with the Hue app.
`
	setControlDocExample = `
# Set the repeat interval of a button, in milliseconds
openhue set control 462e54d9-ec5d-4bf6-879d-ad34cb9a692e --repeat-interval 800

# Same, with a duration
openhue set control 462e54d9-ec5d-4bf6-879d-ad34cb9a692e --repeat-interval 800ms

# Select the button by device name and control number
openhue set control "Hue dimmer switch" --control 2 --repeat-interval 1s

# A short press of button 1 recalls a scene, a second press turns the lights off
openhue set control "Hue dimmer switch" --control 1 --scene Cozy --with-off

# Button 4 cycles through three scenes, restarting after 10 seconds without a press
openhue set control "Hue dimmer switch" --control 4 --scene Soho --scene Nebula --scene Dusk --repeat-timeout 10s

# Holding button 2 dims the lights up, and removing the short-press action of button 3
openhue set control "Hue dimmer switch" --control 2 --on-hold dim-up
openhue set control "Hue dimmer switch" --control 3 --clear

# See what would be sent, without changing anything
openhue set control "Hue dimmer switch" --control 1 --scene Cozy --dry-run`
)

// CmdSetControlOptions holds the settings 'set control' can change: add new fields here as the bridge exposes them.
type CmdSetControlOptions struct {
	control        int
	repeatInterval string
	scenes         []string
	withOff        bool
	repeatTimeout  string
	onHold         string
	clear          bool
	dryRun         bool
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
	cmd.Flags().StringArrayVar(&o.scenes, "scene", nil, "scene (name or ID) recalled by a short press; repeat to cycle through several scenes")
	cmd.Flags().BoolVar(&o.withOff, "with-off", false, "a second press turns the lights off")
	cmd.Flags().StringVar(&o.repeatTimeout, "repeat-timeout", "", "idle time after which a scene cycle restarts: a duration (10s) or a number of seconds")
	cmd.Flags().StringVar(&o.onHold, "on-hold", "", "what holding the button does: dim-up, dim-down, dim-alternate or none")
	cmd.Flags().BoolVar(&o.clear, "clear", false, "remove the short-press action of the button")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "print what would be sent to the bridge, and send nothing")
	_ = cmd.RegisterFlagCompletionFunc("on-hold", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"dim-up", "dim-down", "dim-alternate", "none"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func (o *CmdSetControlOptions) run(ctx *openhue.Context, cmd *cobra.Command, target string) error {

	flags := cmd.Flags()

	// validate the flags before resolving the control, nothing is sent on failure
	behaviorChange := len(o.scenes) > 0 || flags.Changed("with-off") || flags.Changed("repeat-timeout") ||
		flags.Changed("on-hold") || o.clear
	intervalChange := flags.Changed("repeat-interval")
	if !intervalChange && !behaviorChange {
		return fmt.Errorf("no setting to change, available flags: --repeat-interval, --scene, --with-off, --repeat-timeout, --on-hold, --clear")
	}

	interval := 0
	if intervalChange {
		var err error
		if interval, err = parseRepeatInterval(o.repeatInterval); err != nil {
			return err
		}
	}
	patch := openhue.ButtonPatch{OnHold: o.onHold, Clear: o.clear}
	if flags.Changed("with-off") {
		patch.WithOff = &o.withOff
	}
	if flags.Changed("repeat-timeout") {
		var err error
		if patch.RepeatTimeout, err = parseRepeatTimeout(o.repeatTimeout); err != nil {
			return err
		}
	}
	if flags.Changed("on-hold") && o.onHold == "" {
		return fmt.Errorf("--on-hold must be one of dim-up, dim-down, dim-alternate, none")
	}
	if flags.Changed("control") && o.control <= 0 {
		return fmt.Errorf("--control must be a positive control number, got %d", o.control)
	}

	control, err := openhue.FindControl(ctx.Home, target, o.control)
	if err != nil {
		return err
	}

	if control.Kind != openhue.ControlKindButton {
		return fmt.Errorf("control %s is a %s: rotary dials have no configurable settings", control.Id, control.Kind)
	}

	// build the new behavior configuration, still without sending anything
	var behaviorBody *oh.BehaviorInstancePut
	var button map[string]any
	if behaviorChange {
		if control.Behavior == nil {
			return fmt.Errorf("control %s (%s) has no behavior configuration on the bridge: configure the accessory with the Hue app first", control.Id, control.Name)
		}
		if len(o.scenes) > 0 {
			if patch.SceneIds, err = openhue.ControlScenes(ctx.Home, control, o.scenes); err != nil {
				return err
			}
		}
		config, err := openhue.BuildButtonConfiguration(control.Behavior.Configuration, control.Id, patch)
		if err != nil {
			return err
		}
		behaviorBody = &oh.BehaviorInstancePut{Configuration: &config}
		button, _ = config["buttons"].(map[string]any)[control.Id].(map[string]any)
	}

	if o.dryRun {
		if behaviorBody != nil {
			util.PrintJson(ctx.Io, behaviorBody)
		}
		if intervalChange {
			ctx.Io.Printf("Dry run: control '%s' (%s) repeat interval would be set to %dms\n", control.Name, control.Id, interval)
		}
		return nil
	}

	if intervalChange {
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
	}

	if behaviorBody != nil {
		if err := ctx.H.UpdateBehaviorInstance(context.Background(), control.Behavior.Id, *behaviorBody); err != nil {
			cmd.SilenceUsage = true
			return err
		}
		sceneName := func(id string) (string, bool) { return openhue.SceneName(ctx.Home, id) }
		ctx.Io.Printf("Control '%s' (%s) updated: short press: %s, hold: %s\n", control.Name, control.Id,
			openhue.SummarizeShortPress(button, sceneName), openhue.SummarizeHold(button))
	}
	return nil
}

// parseRepeatTimeout accepts a Go duration ("10s") or a plain number of seconds, and returns a duration of at least 1s.
func parseRepeatTimeout(value string) (time.Duration, error) {
	var d time.Duration
	if secs, err := strconv.Atoi(value); err == nil {
		d = time.Duration(secs) * time.Second
	} else if d, err = time.ParseDuration(value); err != nil {
		return 0, fmt.Errorf("invalid --repeat-timeout '%s': use a duration (10s) or a number of seconds", value)
	}
	if d < time.Second {
		return 0, fmt.Errorf("invalid --repeat-timeout '%s': must be at least 1s", value)
	}
	return d, nil
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
