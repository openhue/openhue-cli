package setup

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
	"openhue-cli/openhue/gen"
	"os"
)

type CmdAuthOptions struct {
	bridge            string
	deviceType        string
	generateClientKey bool
}

// NewCmdAuth creates the auth command
func NewCmdAuth(streams openhue.IOStreams) *cobra.Command {

	o := CmdAuthOptions{}

	cmd := &cobra.Command{
		Use:     "auth",
		GroupID: "config",
		Short:   "Retrieve the Hue Application Key",
		Long: `Authenticate to retrieve the Hue Application Key. 

  Requires to go and press the button on the bridge.

You can use the 'openhue discover' command to lookup the IP of your bridge.
`,
		Run: func(cmd *cobra.Command, args []string) {
			RunCmdAuth(streams, o.bridge, o.deviceType, o.generateClientKey)
		},
	}

	cmd.Flags().StringVarP(&o.bridge, "bridge", "b", "", "Bridge IP (example '192.168.1.23')")
	cmd.MarkFlagRequired("bridge")

	cmd.Flags().StringVarP(&o.deviceType, "devicetype", "d", getHostName(), "Device identifier")

	cmd.Flags().BoolVarP(&o.generateClientKey, "generateclientkey", "k", true, "Generate the client key")

	return cmd
}

func RunCmdAuth(streams openhue.IOStreams, bridge string, deviceType string, generateClientKey bool) {
	client := openhue.NewOpenHueClientNoAuth(bridge)

	body := gen.AuthenticateJSONRequestBody{}
	body.Devicetype = &deviceType
	body.Generateclientkey = &generateClientKey
	resp, err := client.AuthenticateWithResponse(context.Background(), body)
	cobra.CheckErr(err)

	auth := (*resp.JSON200)[0]
	if auth.Error != nil {
		fmt.Fprintln(streams.Out, "\n", *auth.Error.Description)
	} else {
		fmt.Fprintln(streams.Out, "\nYour hue-application-key ->", *auth.Success.Username)
	}
}

func getHostName() string {
	hostname, err := os.Hostname()
	cobra.CheckErr(err)
	return hostname
}
