package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"openhue-cli/cmd/get"
	"openhue-cli/cmd/set"
	"openhue-cli/cmd/setup"
	"openhue-cli/cmd/version"
	"openhue-cli/openhue"
	"openhue-cli/util"
)

type OpenHueCmdGroup string

const (
	// OpenHueCmdGroupHue contains the actual commands to control the home
	OpenHueCmdGroupHue = OpenHueCmdGroup("hue")
	// OpenHueCmdGroupConfig contains the commands to configure the CLI
	OpenHueCmdGroupConfig = OpenHueCmdGroup("config")
)

// NewCmdOpenHue represents the `openhue` base command, AKA entry point of the CLI
func NewCmdOpenHue(ctx *openhue.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "openhue",
		Short: "openhue controls your Philips Hue lighting system",
		Long: `
openhue controls your Philips Hue lighting system

    Find more information at: https://www.openhue.io/cli`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Infof("Running the '%s' command", cmd.Name())
			LoadHomeIfNeeded(ctx, cmd)
		},
	}

	cmd.AddGroup(&cobra.Group{
		ID:    string(OpenHueCmdGroupConfig),
		Title: "Configuration",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    string(OpenHueCmdGroupHue),
		Title: "Philips Hue",
	})

	return cmd
}

// LoadHomeIfNeeded checks if the command requires loading the openhue.Home context
func LoadHomeIfNeeded(ctx *openhue.Context, cmd *cobra.Command) {
	if OpenHueCmdGroupHue.containsCmd(cmd) {
		log.Infof("The '%s' command is in the '%s' group so we are loading the Home Context", cmd.Name(), OpenHueCmdGroupHue)
		timer := util.NewTimer()
		home, err := openhue.LoadHome(ctx.Api)
		cobra.CheckErr(err)
		ctx.Home = home
		log.Infof("It took %dms to load the Home Context", timer.SinceInMillis())
	}
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

	// create the root command
	root := NewCmdOpenHue(ctx)

	// add sub commands
	root.AddCommand(version.NewCmdVersion(ctx))
	root.AddCommand(setup.NewCmdSetup(ctx.Io))
	root.AddCommand(setup.NewCmdDiscover(ctx.Io))
	root.AddCommand(setup.NewCmdConfigure(ctx.Io))

	root.AddCommand(set.NewCmdSet(ctx))
	root.AddCommand(get.NewCmdGet(ctx))

	// execute the root command
	err := root.Execute()
	cobra.CheckErr(err)
}

// containsCmd verifies if the given cobra.Command is contained in the group.
func (g OpenHueCmdGroup) containsCmd(cmd *cobra.Command) bool {

	if cmd.GroupID == string(g) {
		return true
	} else if cmd.Parent() != nil {
		return g.containsCmd(cmd.Parent())
	}

	return false
}
