package get

import (
	"fmt"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
	"openhue-cli/util"
)

const (
	docLongGetRoom = `
Fetches and displays all available rooms from the Philips Hue bridge
`
	docExampleGetRoom = `
# List all rooms as a table
openhue get room

# List all rooms as JSON
openhue get room --json

# Get details for a single room
openhue get room aa31ba26-98a7-4830-8ae9-1b7caa8b5700 --json

# Get room by name
openhue get room -n "Studio"
`
)

// NewCmdGetRoom returns initialized Command instance for the 'get light' sub command
func NewCmdGetRoom(ctx *openhue.Context, o *CmdGetOptions) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "room [roomId]",
		Aliases: []string{"rooms"},
		Short:   "Get room",
		Long:    docLongGetRoom,
		Example: docExampleGetRoom,
		Args:    cobra.MatchAll(cobra.RangeArgs(0, 1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			RunGetRoomCmd(ctx, o, args)
		},
	}

	return cmd
}

func RunGetRoomCmd(ctx *openhue.Context, o *CmdGetOptions, args []string) {

	rooms := openhue.SearchRooms(ctx.Home, args)

	if o.Json {
		util.PrintJsonArray(ctx.Io, rooms)
	} else {
		util.PrintTable(ctx.Io, rooms, PrintRoom, "ID", "Name", "Type", "Status", "Brightness")
	}
}

func PrintRoom(room openhue.Room) string {

	status := "[  ]"
	brightness := "N/A"

	if room.GroupedLight.IsOn() {
		status = "[on]"
	}

	if room.GroupedLight.HueData.Dimming != nil {
		brightness = fmt.Sprint(*room.GroupedLight.HueData.Dimming.Brightness) + "%"
	}

	return room.Id + "\t" + room.Name + "\t" + string(*room.HueData.Metadata.Archetype) + "\t" + status + "\t" + brightness
}
