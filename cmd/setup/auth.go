package setup

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"openhue-cli/config"
	"openhue-cli/openhue"
	"os"
)

var (
	bridge            string
	deviceType        string
	generateClientKey bool
)

// NewCmdAuth creates the auth command
func NewCmdAuth() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "auth",
		GroupID: "config",
		Short:   "Retrieve the Hue Application Key",
		Long:    `Authenticate to retrieve the Hue Application Key. Requires to go and press the button on the bridge`,
		Run: func(cmd *cobra.Command, args []string) {

			client := config.NewOpenHueClientNoAuth(bridge)

			body := new(openhue.AuthenticateJSONRequestBody)
			body.Devicetype = &deviceType
			body.Generateclientkey = &generateClientKey
			resp, err := client.AuthenticateWithResponse(context.Background(), *body)
			cobra.CheckErr(err)

			auth := (*resp.JSON200)[0]
			if auth.Error != nil {
				fmt.Println("\n🖲️", *auth.Error.Description)
			} else {
				fmt.Println("\nYour hue-application-key ->", *auth.Success.Username)
			}
		},
	}

	cmd.Flags().StringVarP(&bridge, "bridge", "b", "", "Bridge IP (example '192.168.1.23')")
	cmd.MarkFlagRequired("bridge")

	hostname, err := os.Hostname()
	cobra.CheckErr(err)
	cmd.Flags().StringVarP(&deviceType, "devicetype", "d", hostname, "Device identifier")

	cmd.Flags().BoolVarP(&generateClientKey, "generateclientkey", "k", true, "Generate the client key")

	return cmd
}
