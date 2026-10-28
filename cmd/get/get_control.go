package get

import (
	"fmt"
	"openhue-cli/openhue"
	"openhue-cli/util"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

type ControlOptions struct {
	*CmdGetOptions
	Room   string
	Device string
	Kind   string
}

// NewCmdGetControl returns initialized Command instance for the 'get control' sub command
func NewCmdGetControl(ctx *openhue.Context, co *CmdGetOptions) *cobra.Command {

	o := ControlOptions{
		CmdGetOptions: co,
	}

	cmd := &cobra.Command{
		Use:     "control [controlId]",
		Aliases: []string{"controls"},
		Short:   "Get controls (buttons and rotary dials)",
		Long: `
Displays the physical controls of your switches, dimmers and dials: buttons and rotary dials,
with their device, room, control number on the device and last reported event. For buttons, the
"Short press" and "Hold" columns show what the button is configured to do (see 'openhue set control'):
'scene <name>', 'cycle: <name>, <name>, ... (<n>)', 'time-based', 'smart-scene <id>', 'dim-up', '-' when
nothing is assigned. With --json, each button also carries its raw 'configuration' as stored on the bridge.
`,
		Example: `
# List all controls
openhue get control

# List all controls as JSON
openhue get controls --json

# List the controls of a room
openhue get control --room "Living Room"

# List the controls of a device (name or ID)
openhue get control --device "Hue dimmer switch"

# List only the rotary dials of a room
openhue get control --room "Living Room" --kind rotary

# List specific controls by ID
openhue get control 462e54d9-ec5d-4bf6-879d-ad34cb9a692e`,
		Args: cobra.MatchAll(cobra.RangeArgs(0, 100), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.RunGetControlCmd(ctx, args)
		},
	}

	cmd.Flags().StringVarP(&o.Room, "room", "r", "", "Filter controls by room (name or ID)")
	cmd.Flags().StringVarP(&o.Device, "device", "d", "", "Filter controls by device (name or ID)")
	cmd.Flags().StringVar(&o.Kind, "kind", "", "Filter controls by kind: button or rotary")
	_ = cmd.RegisterFlagCompletionFunc("kind", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"button", "rotary"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func (o *ControlOptions) RunGetControlCmd(ctx *openhue.Context, args []string) error {

	kind := openhue.ControlKind(o.Kind)
	if kind != "" && kind != openhue.ControlKindButton && kind != openhue.ControlKindRotary {
		return fmt.Errorf("invalid --kind '%s', allowed values: button, rotary", o.Kind)
	}

	controls := openhue.SearchControls(ctx.Home, o.Room, o.Device, kind, args)

	// unknown IDs are an error
	for _, id := range args {
		found := false
		for _, c := range controls {
			if c.Id == id {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("control '%s' not found", id)
		}
	}

	if o.Json {
		util.PrintJson(ctx.Io, controls)
	} else {
		util.PrintTable(ctx.Io, controls, PrintControl, "ID", "Kind", "Device", "Room", "Control", "Short press", "Hold", "Last Event", "Time")
	}
	return nil
}

func PrintControl(c openhue.Control) string {

	device, room := "", ""
	if d := c.Device(); d != nil {
		device = d.Name
	}
	if r := c.Room(); r != nil {
		room = r.Name
	}

	event, at := c.LastEvent()
	when := ""
	if at != nil {
		when = at.Local().Format(time.DateTime)
	}

	shortPress, hold := "-", "-"
	if c.Kind == openhue.ControlKindButton {
		shortPress, hold = c.ShortPressSummary(), c.HoldSummary()
	}

	return c.Id +
		"\t" + string(c.Kind) +
		"\t" + device +
		"\t" + room +
		"\t" + strconv.Itoa(c.ControlId) +
		"\t" + shortPress +
		"\t" + hold +
		"\t" + event +
		"\t" + when
}
