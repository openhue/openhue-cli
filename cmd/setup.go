package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:     "setup",
	GroupID: "init",
	Short:   "Configure your local Philips HUE environment",
	Long: `The setup command must be run as a prerequisite for all resource related commands (controlling lights, rooms, scenes, etc.)

It allows to store your Philips Hue Bridge IP and application key in the configuration file (~/.openhue/config.yaml).`,
	Run: Setup,
}

func Setup(cmd *cobra.Command, args []string) {
	err := viper.SafeWriteConfig()
	if err != nil {
		err := viper.WriteConfig()
		cobra.CheckErr(err)
	}
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().StringP("bridge", "b", "", "The local IP of your Philips Hue Bridge (example '192.168.1.68')")
	_ = setupCmd.MarkFlagRequired("bridge")
	_ = viper.BindPFlag("bridge", setupCmd.Flags().Lookup("bridge"))

	setupCmd.Flags().StringP("key", "k", "", "Your Hue Application Key")
	_ = setupCmd.MarkFlagRequired("key")
	_ = viper.BindPFlag("key", setupCmd.Flags().Lookup("key"))
}
