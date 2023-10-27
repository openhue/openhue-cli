package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

var (
	deviceType        string
	generateClientKey bool
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:     "auth",
	GroupID: "init",
	Short:   "Authenticate",
	Long:    `Authenticate to retrieve the HUE application key. Requires to go and press the button on the bridge`,
	Run: func(cmd *cobra.Command, args []string) {

		body := new(openhue.AuthenticateJSONRequestBody)
		body.Devicetype = &deviceType
		body.Generateclientkey = &generateClientKey
		resp, err := openhue.Api.AuthenticateWithResponse(context.Background(), *body)
		cobra.CheckErr(err)

		auth := (*resp.JSON200).Success
		fmt.Println(auth.Clientkey)
	},
}

func init() {
	rootCmd.AddCommand(authCmd)

	authCmd.Flags().StringVarP(&deviceType, "devicetype", "d", "", "Device identifier (example 'app_name#instance_name')")
	authCmd.MarkFlagRequired("devicetype")

	authCmd.Flags().BoolVarP(&generateClientKey, "generateclientkey", "k", true, "")
}
