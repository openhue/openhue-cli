package cmd

import (
	"github.com/spf13/cobra"
)

// roomsCmd represents the rooms command
var roomsCmd = &cobra.Command{
	GroupID: "hue",
	Use:     "rooms",
	Short:   "Control rooms",
	Long:    `Control the rooms available from the Philips HUE bridge`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(roomsCmd)
}
