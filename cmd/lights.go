package cmd

import (
	"github.com/spf13/cobra"
)

// lightsCmd represents the lights command
var lightsCmd = &cobra.Command{
	GroupID: "hue",
	Use:     "lights",
	Short:   "Control lights",
	Long:    `Control all available lights`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(lightsCmd)
}
