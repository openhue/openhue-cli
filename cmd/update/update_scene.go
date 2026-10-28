package update

import (
	"fmt"

	"github.com/spf13/cobra"
	"openhue-cli/cmd/scenecmd"
	"openhue-cli/openhue"
	"openhue-cli/util"
)

const (
	updateSceneDocShort = "Update a scene"
	updateSceneDocLong  = `
Edits an existing scene. Changes apply on top of the values stored in the scene (not the current state of the
lights), so the lights you do not mention keep their settings, including effects and gradients. The real
lights are never changed.

A --light for a light of the room that is not yet in the scene adds it, starting from its current state.
--from-file is applied as a patch: what it does not mention is left untouched.

See 'openhue create scene --help' for the --light settings.
`
	updateSceneDocExample = `
# Change one light of a scene
openhue update scene Cozy --light "Desk:brightness=20"

# Dim all the lights of a scene
openhue update scene Cozy --brightness 10

# Add a light, turn another one off in the scene
openhue update scene Cozy --light "Shelf:on,brightness=30" --remove-light Ceiling

# Rename a scene
openhue update scene Cozy --name Relax

# Change the speed of a dynamic scene, or clear its palette
openhue update scene Aurora --speed 60
openhue update scene Aurora --clear-palette

# Apply an edited export
openhue get scene Cozy -r "Living Room" --yaml > cozy.yaml
openhue update scene Cozy -r "Living Room" --from-file cozy.yaml

# Print the merged request without sending it
openhue update scene Cozy --brightness 10 --dry-run`
)

type sceneOptions struct {
	scenecmd.Flags
}

// NewCmdUpdateScene returns initialized cobra.Command instance for the 'update scene' sub command
func NewCmdUpdateScene(ctx *openhue.Context) *cobra.Command {

	o := &sceneOptions{}

	cmd := &cobra.Command{
		Use:           "scene [sceneId|sceneName]",
		Aliases:       []string{"scenes"},
		Short:         updateSceneDocShort,
		Long:          updateSceneDocLong,
		Example:       updateSceneDocExample,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(ctx, cmd, args)
		},
	}

	o.Register(cmd, true)

	return cmd
}

func (o *sceneOptions) run(ctx *openhue.Context, cmd *cobra.Command, args []string) error {

	scene, room, err := openhue.FindScene(ctx.Home, o.Room, args[0])
	if err != nil {
		return err
	}

	def, err := o.Resolve(cmd, ctx.Io.In, "", false)
	if err != nil {
		return err
	}
	if def.Room != "" && def.Room != room.Name && def.Room != room.Id {
		return fmt.Errorf("the room '%s' of the definition does not match the room '%s' of the scene", def.Room, room.Name)
	}
	if def.IsEmpty() && def.Name == "" {
		return fmt.Errorf("at least one change is required (see 'openhue update scene --help')")
	}

	put, warnings, err := openhue.BuildScenePut(room, scene, def)
	if err != nil {
		return err
	}
	scenecmd.PrintWarnings(ctx.Io, warnings)

	if o.DryRun {
		return scenecmd.PrintRequest(ctx.Io, put)
	}

	if err := scene.Update(ctx.H, *put); err != nil {
		return err
	}

	if o.Json {
		updated, err := openhue.ReloadScene(ctx.H, scene.Id)
		if err != nil {
			return err
		}
		util.PrintJson(ctx.Io, updated)
		return nil
	}

	name := scene.Name
	if put.Metadata != nil && put.Metadata.Name != nil {
		name = *put.Metadata.Name
	}
	ctx.Io.Printf("Scene '%s' updated in room '%s' (%s)\n", name, room.Name, scene.Id)
	return nil
}
