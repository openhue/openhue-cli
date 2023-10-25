package cmd

import (
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "openhue-cli",
	Short: "openhue-cli controls your Philips Hue lighting system",
	Long: `openhue-cli controls your Philips Hue lighting system

    Find more information at: https://www.openhue.io/cli`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func init() {
	cobra.OnInitialize(openhue.Init)

	rootCmd.AddGroup(&cobra.Group{
		ID:    "init",
		Title: "Initialization",
	})

	rootCmd.AddGroup(&cobra.Group{
		ID:    "hue",
		Title: "Philips HUE",
	})
}
