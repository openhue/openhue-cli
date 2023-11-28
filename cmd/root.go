package cmd

import (
	"github.com/spf13/cobra"
	"openhue-cli/cmd/get"
	"openhue-cli/cmd/set"
	"openhue-cli/cmd/setup"
	"openhue-cli/cmd/version"
	"openhue-cli/openhue"
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
func Execute(buildInfo *openhue.BuildInfo) {

	// load the configuration
	c := openhue.Config{}
	c.LoadConfig()

	// get the API Client
	api := c.NewOpenHueClient()
	ctx := openhue.NewContext(openhue.NewIOSteams(), buildInfo, api)

	// create the root command
	root := NewCmdOpenHue()

	// init groups
	initGroups(root)

	// add sub commands
	root.AddCommand(version.NewCmdVersion(ctx))
	root.AddCommand(setup.NewCmdAuth())
	root.AddCommand(setup.NewCmdDiscover())
	root.AddCommand(setup.NewCmdConfigure())
	root.AddCommand(set.NewCmdSet(ctx))
	root.AddCommand(get.NewCmdGet(ctx))

	// execute the root command
	err := root.Execute()
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
