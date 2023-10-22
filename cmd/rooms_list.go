package cmd

import (
	"context"
	"fmt"
	"openhue-cli/openhue"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listRoomsCmd = &cobra.Command{
	Use:   "list",
	Short: "Display all rooms",
	Long:  `Fetches and displays all available rooms from the Philips Hue bridge`,
	Run: func(cmd *cobra.Command, args []string) {

		resp, err := openhue.Api.GetRoomsWithResponse(context.Background())
		cobra.CheckErr(err)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "Room ID\tRoom Name")
		_, _ = fmt.Fprintln(w, "----\t----")

		rooms := (*resp.JSON200).Data
		for _, l := range *rooms {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", *l.Id, *l.Metadata.Name)
		}

		_ = w.Flush()
	},
}

func init() {
	roomsCmd.AddCommand(listRoomsCmd)
}
