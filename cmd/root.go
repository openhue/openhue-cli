package cmd

import (
	"github.com/spf13/cobra"
	"openhue-cli/cmd/get"
	"openhue-cli/cmd/set"
	"openhue-cli/cmd/setup"
	"openhue-cli/cmd/version"
	"openhue-cli/config"
	"openhue-cli/util"
)

// NewCmdOpenHue represents the `openhue` base command, AKA entry point of the CLI
func NewCmdOpenHue() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "openhue",
		Short: "openhue controls your Philips Hue lighting system",
		Long: `
openhue controls your Philips Hue lighting system

    Find more information at: https://www.openhue.io/cli`,
	}

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(buildInfo *util.BuildInfo) {

	// load the configuration
	c := config.Config{}
	c.Load()

	// get the API Client
	api := c.NewOpenHueClient()

	// create the root command
	cmd := NewCmdOpenHue()

	// init groups
	initGroups(cmd)

	// add sub commands
	cmd.AddCommand(version.NewCmdVersion(buildInfo))
	cmd.AddCommand(setup.NewCmdAuth())
	cmd.AddCommand(setup.NewCmdDiscover())
	cmd.AddCommand(setup.NewCmdConfigure())
	cmd.AddCommand(set.NewCmdSet(api))
	cmd.AddCommand(get.NewCmdGet(api))

	// execute the root command
	err := cmd.Execute()
	cobra.CheckErr(err)
}

func initGroups(rootCmd *cobra.Command) {
	rootCmd.AddGroup(&cobra.Group{
		ID:    "config",
		Title: "Configuration",
	})

	rootCmd.AddGroup(&cobra.Group{
		ID:    "hue",
		Title: "Philips Hue",
	})
}
