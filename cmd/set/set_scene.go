package set

import (
	"errors"
	oh "github.com/openhue/openhue-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

const (
	setSceneDocShort = "Activate a scene"
	setSceneDocLong  = `
This command allows the activate a scene using either a static or a dynamic palette.
`
	setSceneDocExample = `
# Activate a scene
openhue set scene Soho

# Activate a scene by ID
openhue set scene 62af7df3-d390-4408-a7ac-4b6b8805531b

# Activate a scene in a specific room
openhue set scene Soho -r Studio

# Activate a scene with a dynamic palette
openhue set scene Soho -a dynamic`
)

type CmdSetSceneOptions struct {
	action actionEnum
	room   string
}

// NewCmdSetScene returns initialized cobra.Command instance for the 'set scene' sub command
func NewCmdSetScene(ctx *openhue.Context) *cobra.Command {

	o := &CmdSetSceneOptions{}

	cmd := &cobra.Command{
		Use:     "scene [sceneId|sceneName]",
		Aliases: []string{"scenes"},
		Short:   setSceneDocShort,
		Long:    setSceneDocLong,
		Example: setSceneDocExample,
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {

			scenes := openhue.SearchScenes(ctx.Home, o.room, args)

			if len(scenes) == 0 {
				ctx.Io.ErrPrintf("Scene '%s' not found%s\n", args[0], o.roomMsg())
				return
			}

			for _, scene := range scenes {
				err := scene.Activate(o.action.toSceneRecallAction())
				if err != nil {
					ctx.Io.ErrPrintln(err.Error())
					continue
				}
				ctx.Io.Printf("Scene '%s' activated%s\n", scene.Name, o.roomMsg())
			}
		},
	}

	cmd.Flags().StringVarP(&o.room, "room", "r", "", "room where the scene is located (in case multiple scenes with the same name exist)")

	// action flag
	cmd.Flags().VarP(&o.action, "action", "a", "action to perform on the scene. allowed: active (default), dynamic or static")
	_ = cmd.RegisterFlagCompletionFunc("action", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"active\tthe actions in the scene are executed on the target",
			"static\t(default)",
			"dynamic\tstarts dynamic scene with colors in the Palette object",
		}, cobra.ShellCompDirectiveDefault
	})

	return cmd
}

type actionEnum string

const (
	active  actionEnum = "active"
	static  actionEnum = "static"
	dynamic actionEnum = "dynamic"
)

// String is used both by fmt.Print and by Cobra in help text
func (e *actionEnum) String() string {
	return string(*e)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *actionEnum) Set(v string) error {
	switch v {
	case "active", "static", "dynamic":
		*e = actionEnum(v)
		return nil
	default:
		return errors.New(`must be one of "active", "static", or "dynamic"`)
	}
}

// Type is only used in help text
func (e *actionEnum) Type() string {
	return "string"
}

func (e *actionEnum) toSceneRecallAction() oh.SceneRecallAction {

	if len(string(*e)) == 0 {
		log.Info("action flag not set, defaulting to 'active")
		*e = "active"
	}

	if *e == dynamic {
		return oh.SceneRecallActionDynamicPalette
	}
	return oh.SceneRecallAction(e.String())
}

func (o *CmdSetSceneOptions) roomMsg() string {
	if o.room != "" {
		return " in room '" + o.room + "'"
	}
	return ""
}
