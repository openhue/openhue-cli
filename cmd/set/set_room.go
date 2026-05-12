package set

import (
	"openhue-cli/openhue"

	"github.com/spf13/cobra"
)

const (
	setRoomDocShort = "Update one or multiple rooms (on/off, brightness, color)"
	setRoomDocLong  = `
Update one or multiple rooms (max is 10 rooms simultaneously).

 You must set the --on flag in order to turn a room on, even if you set the brightness or the color values.
    
Use "openhue get room" for a complete list of available rooms.
`
	setRoomDocExample = `
# Turn on a room
openhue set room 15f51223-1e83-4e48-9158-0c20dbd5734e --on

# Turn on multiple rooms
openhue set room 83111103-a3eb-40c5-b22a-02deedd21fcb 8f0a7b52-df25-4bc7-b94d-0dd1a88068ff --on

# Turn off a room identified by its name
openhue set room Studio --off

# Set brightness of a single room with transition time
openhue set room 15f51223-1e83-4e48-9158-0c20dbd5734e --on --brightness 42.65 --transition-time 5s

# Set color (in RGB) of a single room
openhue set room 15f51223-1e83-4e48-9158-0c20dbd5734e --on --rgb #3399FF

# Set color (in CIE space) of a single room
openhue set room 15f51223-1e83-4e48-9158-0c20dbd5734e --on -x 0.675 -y 0.322

# Set color (by name) of a single room
openhue set room 15f51223-1e83-4e48-9158-0c20dbd5734e --on --color powder_blue

# Set color temperature (in Mirek) of a single room
openhue set light Studio --on -t 250
`
)

// NewCmdSetRoom returns initialized cobra.Command instance for the 'set room' sub command
func NewCmdSetRoom(ctx *openhue.Context) *cobra.Command {

	f := CmdSetLightFlags{}

	cmd := &cobra.Command{
		Use:     "room [roomId]",
		Aliases: []string{"rooms"},
		Short:   setRoomDocShort,
		Long:    setRoomDocLong,
		Example: setRoomDocExample,
		Args:    cobra.MatchAll(cobra.RangeArgs(1, 10), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			o, err := f.toSetLightOptions()
			cobra.CheckErr(err)

			rooms := openhue.SearchRooms(ctx.Home, args)

			if len(rooms) == 0 {
				ctx.Io.ErrPrintln("no room(s) found for given ID(s)", args)
			}

			for _, room := range rooms {
				if room.GroupedLight == nil {
					ctx.Io.ErrPrintln("room", room.Name, "has no grouped light service")
					continue
				}
				if err := room.GroupedLight.Set(o); err != nil {
					ctx.Io.ErrPrintln("failed to set room", room.Name+":", err)
				}
			}
		},
	}

	f.initCmd(cmd)

	return cmd
}
