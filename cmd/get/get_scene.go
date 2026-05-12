package get

import (
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
	"openhue-cli/util"
	"strconv"
)

type SceneOptions struct {
	*CmdGetOptions
	Room string
}

// NewCmdGetScene returns initialized Command instance for the 'get scene' sub command
func NewCmdGetScene(ctx *openhue.Context, co *CmdGetOptions) *cobra.Command {

	o := SceneOptions{
		CmdGetOptions: co,
	}

	cmd := &cobra.Command{
		Use:     "scene [sceneId|sceneName]",
		Aliases: []string{"scenes"},
		Short:   "Get scenes",
		Long: `
Displays all the scenes with their main information, including the room they belong to.
`,
		Example: `
# List all scenes
openhue get scene

# List all scenes as JSON 
openhue get scene --json

# Filter scenes for a given room name
openhue get scenes --room "Living Room"

# Filter scenes for a given room ID
openhue get scenes -r 878a65d6-613b-4239-8b77-588b535bfb4a

# List multiple scenes using either the ID or the name of the scene
openhue get scenes "Palm Beach" Nebula 462e54d9-ec5d-4bf6-879d-ad34cb9a692e`,
		Args: cobra.MatchAll(cobra.RangeArgs(0, 100), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			o.RunGetSceneCmd(ctx, args)
		},
	}

	cmd.Flags().StringVarP(&o.Room, "room", "r", "", "Filter scenes by room (name or ID)")

	return cmd
}

func (o *SceneOptions) RunGetSceneCmd(ctx *openhue.Context, args []string) {

	var scenes []openhue.Scene

	scenes = openhue.SearchScenes(ctx.Home, o.Room, args)

	if o.Json {
		util.PrintJsonArray(ctx.Io, scenes)
	} else {
		util.PrintTable(ctx.Io, scenes, PrintScene, "ID", "Name", "Room", "Status", "Speed", "Auto Dynamic")
	}
}

func PrintScene(scene openhue.Scene) string {

	parent := "unknown"
	active := "unknown"
	speed := "unknown"
	auto := "unknown"

	if scene.Parent != nil {
		parent = scene.Parent.Name
	}

	if scene.HueData != nil {
		if scene.HueData.Status != nil && scene.HueData.Status.Active != nil {
			active = string(*scene.HueData.Status.Active)
		}
		if scene.HueData.Speed != nil {
			speed = strconv.FormatFloat(float64(*scene.HueData.Speed), 'f', 2, 64)
		}
		if scene.HueData.AutoDynamic != nil {
			auto = strconv.FormatBool(*scene.HueData.AutoDynamic)
		}
	}

	return scene.Id +
		"\t" + scene.Name +
		"\t" + parent +
		"\t" + active +
		"\t" + speed +
		"\t" + auto
}
