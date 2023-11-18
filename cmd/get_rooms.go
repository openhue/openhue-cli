package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"openhue-cli/openhue"
	"os"
)

// roomCmd represents the room command
var roomCmd = &cobra.Command{
	Use:   "room",
	Short: "Get room",
	Long: `
Fetches and displays all available rooms from the Philips Hue bridge
`,
	Args: cobra.MatchAll(cobra.RangeArgs(0, 1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {

		var rooms *[]openhue.RoomGet

		if len(args) > 0 {
			resp, err := openhue.Api.GetRoomWithResponse(context.Background(), args[0])
			cobra.CheckErr(err)

			if resp.JSON200 == nil {
				fmt.Println("\nNot room found with ID", args[0])
				os.Exit(0)
			}

			rooms = (*resp.JSON200).Data
		} else {
			resp, err := openhue.Api.GetRoomsWithResponse(context.Background())
			cobra.CheckErr(err)
			rooms = (*resp.JSON200).Data
		}

		if !GetConfig.Json {
			openhue.PrintTable(*rooms, PrintRoom, "Room ID", "Room Name", "Room Type")
		} else {
			openhue.PrintJsonArray(*rooms)
		}
	},
}

func init() {
	getCmd.AddCommand(roomCmd)
}

func PrintRoom(room openhue.RoomGet) string {
	return *room.Id + "\t" + *room.Metadata.Name + "\t" + string(*room.Metadata.Archetype)
}
