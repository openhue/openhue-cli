package setup

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
	"openhue-cli/openhue"
	"openhue-cli/openhue/gen"
	"openhue-cli/util/mdns"
	"os"
	"time"
)

type CmdSetupOptions struct {
	deviceType        string
	generateClientKey bool
}

func NewCmdSetup(io openhue.IOStreams) *cobra.Command {

	o := CmdSetupOptions{}

	cmd := &cobra.Command{
		Use:     "setup",
		GroupID: "config",
		Short:   "Automatic openhue CLI setup",
		Long: `
The setup command will automatically discover the Hue Bridge connected to your local network and ask you to push 
the bridge button to perform initial pairing.
`,
		Run: func(cmd *cobra.Command, args []string) {
			startSetup(io, &o)
		},
	}

	cmd.Flags().StringVarP(&o.deviceType, "devicetype", "d", getHostName(), "Device identifier")
	cmd.Flags().BoolVarP(&o.generateClientKey, "generateclientkey", "k", true, "Generate the client key")

	return cmd
}

func startSetup(io openhue.IOStreams, o *CmdSetupOptions) {
	ipChan := make(chan *net.IP)
	go mdns.DiscoverBridge(ipChan, 5*time.Second)
	ip := <-ipChan

	if ip == nil {
		fmt.Fprintf(io.ErrOut, "❌ Unable to discover your Hue Bridge on your local network\n")
		return
	}

	fmt.Fprintf(io.Out, "[OK] Found Hue Bridge with IP '%s'\n", ip)

	client := openhue.NewOpenHueClientNoAuth(ip.String())

	fmt.Fprintln(io.Out, "[..] Please push the button on your Hue Bridge")
	done := false
	for done == false {
		fmt.Fprintf(io.Out, ".")
		key, err := tryAuth(client, o.toAuthenticateBody())
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		done = true
		fmt.Fprintf(io.Out, "\n")
		log.Info("Hue Application Key is ", key)
		fmt.Fprintln(io.Out, "[OK] Successfully paired openhue with your Hue Bridge!")
		path, err := saveConfig(ip.String(), key)
		if err != nil {
			fmt.Fprintf(io.ErrOut, "[KO] Unable to save config")
		}
		fmt.Fprintln(io.Out, "[OK] Configuration saved in file", path)
	}
}

func (o *CmdSetupOptions) toAuthenticateBody() gen.AuthenticateJSONRequestBody {
	body := gen.AuthenticateJSONRequestBody{}
	body.Devicetype = &o.deviceType
	body.Generateclientkey = &o.generateClientKey
	return body
}

func tryAuth(client *gen.ClientWithResponses, body gen.AuthenticateJSONRequestBody) (string, error) {

	resp, err := client.AuthenticateWithResponse(context.Background(), body)
	cobra.CheckErr(err)

	auth := (*resp.JSON200)[0]
	if auth.Error != nil {
		return "", errors.New(*auth.Error.Description)
	}

	return *auth.Success.Username, nil
}

func saveConfig(bridge string, key string) (string, error) {
	c := openhue.Config{
		Bridge: bridge,
		Key:    key,
	}

	return c.Save()
}

func getHostName() string {
	hostname, err := os.Hostname()
	cobra.CheckErr(err)
	return hostname
}
