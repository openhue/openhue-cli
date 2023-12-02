package setup

import (
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

const (
	docShortConfigure = `Configure your local Philips Hue environment`
	docLongConfigure  = `
The setup command must be run as a prerequisite for all resource related commands (controlling lights, rooms, scenes, etc.)

It allows to store your Philips Hue Bridge IP and application key in the configuration file (~/.openhue/config.yaml).`
)

type Options struct {
	bridge string
	key    string
}

// NewCmdConfigure creates the configure command
func NewCmdConfigure() *cobra.Command {

	o := Options{}

	cmd := &cobra.Command{
		Use:     "configure",
		GroupID: "config",
		Short:   docShortConfigure,
		Long:    docLongConfigure,
		Run: func(cmd *cobra.Command, args []string) {

			c := openhue.Config{
				Bridge: o.bridge,
				Key:    o.key,
			}

			err := c.Save()
			cobra.CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&o.bridge, "bridge", "b", "", "The local IP of your Philips Hue Bridge (example '192.168.1.68')")
	_ = cmd.MarkFlagRequired("bridge")

	cmd.Flags().StringVarP(&o.key, "key", "k", "", "Your Hue Application Key")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}
