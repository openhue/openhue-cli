package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
	"os"
)

var (
	bridge            string
	deviceType        string
	generateClientKey bool
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:     "auth",
	GroupID: "init",
	Short:   "Authenticate",
	Long:    `Authenticate to retrieve the Hue Application Key. Requires to go and press the button on the bridge`,
	Run: func(cmd *cobra.Command, args []string) {

		client := openhue.NewOpenHueClientNoAuth(bridge)

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

func init() {
	rootCmd.AddCommand(authCmd)

	authCmd.Flags().StringVarP(&bridge, "bridge", "b", "", "Bridge IP (example '192.168.1.23')")
	authCmd.MarkFlagRequired("bridge")

	hostname, err := os.Hostname()
	cobra.CheckErr(err)
	authCmd.Flags().StringVarP(&deviceType, "devicetype", "d", hostname, "Device identifier")

	authCmd.Flags().BoolVarP(&generateClientKey, "generateclientkey", "k", true, "Generate the client key")
}
