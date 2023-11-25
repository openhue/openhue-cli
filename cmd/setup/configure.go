package setup

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	docShortConfigure = `Configure your local Philips Hue environment`
	docLongConfigure  = `
The setup command must be run as a prerequisite for all resource related commands (controlling lights, rooms, scenes, etc.)

It allows to store your Philips Hue Bridge IP and application key in the configuration file (~/.openhue/config.yaml).`
)

// NewCmdConfigure creates the configure command
func NewCmdConfigure() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "configure",
		GroupID: "config",
		Short:   docShortConfigure,
		Long:    docLongConfigure,
		Run: func(cmd *cobra.Command, args []string) {
			err := viper.SafeWriteConfig()
			if err != nil {
				err := viper.WriteConfig()
				cobra.CheckErr(err)
			}
		},
	}

	cmd.Flags().StringP("bridge", "b", "", "The local IP of your Philips Hue Bridge (example '192.168.1.68')")
	_ = cmd.MarkFlagRequired("bridge")
	_ = viper.BindPFlag("bridge", cmd.Flags().Lookup("bridge"))

	cmd.Flags().StringP("key", "k", "", "Your Hue Application Key")
	_ = cmd.MarkFlagRequired("key")
	_ = viper.BindPFlag("key", cmd.Flags().Lookup("key"))

	return cmd
}
