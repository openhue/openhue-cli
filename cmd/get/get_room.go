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

type RoomOptions struct {
	RoomParam string
	Json      *bool
	Name      *bool
}

func NewGetRoomOptions(co *CmdGetOptions) *RoomOptions {
	return &RoomOptions{
		Json: &co.Json,
		Name: &co.Name,
	}
}

// NewCmdGetRoom returns initialized Command instance for the 'get light' sub command
func NewCmdGetRoom(ctx *openhue.Context, co *CmdGetOptions) *cobra.Command {

	o := NewGetRoomOptions(co)

	cmd := &cobra.Command{
		Use:     "room [roomId]",
		Short:   "Get room",
		Long:    docLongGetRoom,
		Example: docExampleGetRoom,
		Args:    cobra.MatchAll(cobra.RangeArgs(0, 1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			o.PrepareGetRoomCmd(args)
			o.RunGetRoomCmd(ctx)
		},
	}

	return cmd
}

func (o *RoomOptions) PrepareGetRoomCmd(args []string) {
	if len(args) > 0 {
		o.RoomParam = args[0]
	}
}

func (o *RoomOptions) RunGetRoomCmd(ctx *openhue.Context) {

	var rooms []openhue.Room

	if len(o.RoomParam) > 0 {

		if *o.Name {
			rooms = openhue.FindRoomByName(ctx.Home, o.RoomParam)
		} else {
			rooms = openhue.FindRoomById(ctx.Home, o.RoomParam)
		}

	} else {
		rooms = openhue.FindAllRooms(ctx.Home)
	}

	if *o.Json {
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
