package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"openhue-cli/cmd/get"
	"openhue-cli/cmd/set"
	"openhue-cli/cmd/setup"
	"openhue-cli/cmd/version"
	"openhue-cli/openhue"
	"os"
	"time"
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
	c.Load()

	// get the API Client
	api := c.NewOpenHueClient()
	ctx := openhue.NewContext(openhue.NewIOStreams(), buildInfo, api)

	// load the home context
	t0 := time.Now()
	home, err := openhue.LoadHome(api)
	cobra.CheckErr(err)
	ctx.Home = home
	log.Infof("It took %dms to load the Home Context", time.Since(t0).Milliseconds())

	// create the root command
	root := NewCmdOpenHue()
	log.Infof("Running the '%s' command", os.Args)

	// init groups
	initGroups(root)

	// add sub commands
	root.AddCommand(version.NewCmdVersion(ctx))
	root.AddCommand(setup.NewCmdAuth(ctx.Io))
	root.AddCommand(setup.NewCmdDiscover(ctx.Io))
	root.AddCommand(setup.NewCmdConfigure())

	root.AddCommand(set.NewCmdSet(ctx))
	root.AddCommand(get.NewCmdGet(ctx))

	// execute the root command
	err = root.Execute()
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
