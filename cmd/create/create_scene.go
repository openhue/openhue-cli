package create

import (
	"fmt"

	"github.com/spf13/cobra"
	"openhue-cli/cmd/scenecmd"
	"openhue-cli/openhue"
	"openhue-cli/util"
)

const (
	createSceneDocShort = "Create a scene"
	createSceneDocLong  = `
Creates a scene in a room. By default the scene stores the current state of every light of the room.
The real lights are never changed.

Settings are layered: current state, then the default flags (--on, --brightness, --color, ...), then the
per-light settings (--light "<light>:<settings>"). Flags override the values of --from-file.

Per-light settings are comma-separated: on, off, brightness=<0-100>, rgb=<#hex>, x=<n>, y=<n>, color=<name>,
temperature=<153-500>, transition=<duration>, effect=<name>, gradient=<color>;<color>;..., gradient-mode=<mode>.

The definition file format is the one printed by 'openhue get scene <scene> --yaml'.
`
	createSceneDocExample = `
# Save the current look of a room
openhue create scene Cozy --room "Living Room"

# Define each light
openhue create scene Work -r Office --light "Desk:on,brightness=100,temperature=250" --light "Ceiling:off"

# Dim everything, except one lamp, with a 3 seconds transition
openhue create scene Dim -r Bedroom --brightness 10 --transition-time 3s --light "Lamp:brightness=40"

# Only the listed lights are lit, the other ones are turned off in the scene
openhue create scene Desk -r Office --only-listed --light "Desk:on"

# Dynamic scene
openhue create scene Aurora -r Lounge --auto-dynamic --speed 30 --palette-color blue --palette-color "#00FF88"

# From a definition file, or from stdin
openhue create scene --from-file sunset.yaml
cat sunset.yaml | openhue create scene Sunset2 --from-file -

# Print the request without sending it
openhue create scene Cozy -r "Living Room" --dry-run`
)

type sceneOptions struct {
	scenecmd.Flags
	OnlyListed bool
}

// NewCmdCreateScene returns initialized cobra.Command instance for the 'create scene' sub command
func NewCmdCreateScene(ctx *openhue.Context) *cobra.Command {

	o := &sceneOptions{}

	cmd := &cobra.Command{
		Use:           "scene [name]",
		Aliases:       []string{"scenes"},
		Short:         createSceneDocShort,
		Long:          createSceneDocLong,
		Example:       createSceneDocExample,
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(ctx, cmd, args)
		},
	}

	o.Register(cmd, false)
	cmd.Flags().BoolVar(&o.OnlyListed, "only-listed", false, "Turn off the lights that are not given with --light")

	return cmd
}

func (o *sceneOptions) run(ctx *openhue.Context, cmd *cobra.Command, args []string) error {

	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	def, err := o.Resolve(cmd, ctx.Io.In, name, true)
	if err != nil {
		return err
	}
	if def.Name == "" {
		return fmt.Errorf("a scene name is required (argument or 'name' in the file)")
	}
	if def.Room == "" {
		return fmt.Errorf("a room is required (--room or 'room' in the file)")
	}

	room, err := openhue.FindRoom(ctx.Home, def.Room)
	if err != nil {
		return err
	}

	post, warnings, err := openhue.BuildScenePost(room, def, o.OnlyListed)
	if err != nil {
		return err
	}
	scenecmd.PrintWarnings(ctx.Io, warnings)

	if o.DryRun {
		return scenecmd.PrintRequest(ctx.Io, post)
	}

	id, err := room.CreateScene(ctx.H, *post)
	if err != nil {
		return err
	}

	if o.Json {
		created, err := openhue.ReloadScene(ctx.H, id)
		if err != nil {
			return err
		}
		util.PrintJson(ctx.Io, created)
		return nil
	}

	ctx.Io.Printf("Scene '%s' created in room '%s' (%s)\n", def.Name, room.Name, id)
	return nil
}
