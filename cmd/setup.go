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
	Long: `The setup command must be run as a prerequisite. It allows to store your Philips HUE Bridge IP and 
application key in the configuration file`,
	Run: Setup,
}

func Setup(cmd *cobra.Command, args []string) {
	err := viper.SafeWriteConfig()
	cobra.CheckErr(err)
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.PersistentFlags().StringP("ip", "i", "192.168.1.68", "The local IP of your Philips HUE Bridge")
	_ = setupCmd.MarkPersistentFlagRequired("ip")
	_ = viper.BindPFlag("ip", setupCmd.PersistentFlags().Lookup("ip"))

	setupCmd.PersistentFlags().StringP("key", "k", "####", "Your HUE application key")
	_ = setupCmd.MarkPersistentFlagRequired("key")
	_ = viper.BindPFlag("key", setupCmd.PersistentFlags().Lookup("key"))
}
