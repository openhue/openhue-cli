package get

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
	"openhue-cli/util"
	"os"
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
`
)

type RoomOptions struct {
	RoomId string
}

func NewGetRoomOptions() *RoomOptions {
	return &RoomOptions{}
}

// NewCmdGetRoom returns initialized Command instance for the 'get light' sub command
func NewCmdGetRoom(api *openhue.ClientWithResponses) *cobra.Command {

	o := NewGetRoomOptions()

	cmd := &cobra.Command{
		Use:     "room [roomId]",
		Short:   "Get room",
		Long:    docLongGetRoom,
		Example: docExampleGetRoom,
		Args:    cobra.MatchAll(cobra.RangeArgs(0, 1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			o.PrepareGetRoomCmd(args)
			o.RunGetRoomCmd(api)
		},
	}

	return cmd
}

func (o *RoomOptions) PrepareGetRoomCmd(args []string) {
	if len(args) > 0 {
		o.RoomId = args[0]
	}
}

func (o *RoomOptions) RunGetRoomCmd(api *openhue.ClientWithResponses) {
	var rooms *[]openhue.RoomGet

	if len(o.RoomId) > 0 {
		resp, err := api.GetRoomWithResponse(context.Background(), o.RoomId)
		cobra.CheckErr(err)

		if resp.JSON200 == nil {
			fmt.Println("\nNot room found with ID", o.RoomId)
			os.Exit(0)
		}

		rooms = (*resp.JSON200).Data
	} else {
		resp, err := api.GetRoomsWithResponse(context.Background())
		cobra.CheckErr(err)
		rooms = (*resp.JSON200).Data
	}

	if !GetConfig.Json {
		util.PrintTable(*rooms, PrintRoom, "ID", "Name", "Type")
	} else {
		util.PrintJsonArray(*rooms)
	}
}

func PrintRoom(room openhue.RoomGet) string {
	return *room.Id + "\t" + *room.Metadata.Name + "\t" + string(*room.Metadata.Archetype)
}
