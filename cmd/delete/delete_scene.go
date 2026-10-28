package delete

import (
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

const (
	deleteSceneDocShort = "Delete a scene"
	deleteSceneDocLong  = `
Deletes exactly one scene, identified by its name or its ID. This cannot be undone: use 'openhue get scene'
(optionally with --yaml to keep a copy) beforehand.
`
	deleteSceneDocExample = `
# Delete a scene
openhue delete scene Cozy

# Delete a scene by ID
openhue delete scene 62af7df3-d390-4408-a7ac-4b6b8805531b

# Delete a scene in a specific room (when several rooms have a scene with this name)
openhue delete scene Cozy -r "Living Room"`
)

// NewCmdDeleteScene returns initialized cobra.Command instance for the 'delete scene' sub command
func NewCmdDeleteScene(ctx *openhue.Context) *cobra.Command {

	var room string

	cmd := &cobra.Command{
		Use:           "scene [sceneId|sceneName]",
		Aliases:       []string{"scenes"},
		Short:         deleteSceneDocShort,
		Long:          deleteSceneDocLong,
		Example:       deleteSceneDocExample,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			scene, r, err := openhue.FindScene(ctx.Home, room, args[0])
			if err != nil {
				return err
			}

			if err := scene.Delete(ctx.H); err != nil {
				return err
			}

			ctx.Io.Printf("Scene '%s' deleted from room '%s' (%s)\n", scene.Name, r.Name, scene.Id)
			return nil
		},
	}

	cmd.Flags().StringVarP(&room, "room", "r", "", "room where the scene is located (in case multiple scenes with the same name exist)")

	return cmd
}
