package setup

import (
	"fmt"
	oh "github.com/openhue/openhue-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
	"os"
	"time"
)

const hueBridgeDiscover = "https://discovery.meethue.com"

type CmdSetupOptions struct {
	deviceType        string
	generateClientKey bool
	bridge            string
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
	cmd.Flags().StringVarP(&o.bridge, "bridge", "b", "", "Set your Hue Bridge IP address ("+hueBridgeDiscover+")")

	return cmd
}

func startSetup(io openhue.IOStreams, o *CmdSetupOptions) {
	ip, err := getBridgeIPAddress(io, o)
	if err != nil {
		io.ErrPrintln(err.Error())
		return
	}

	authenticator, err := oh.NewAuthenticator(ip, oh.WithDeviceType(o.deviceType), oh.WithGenerateClientKey(o.generateClientKey))
	if err != nil {
		io.ErrPrintln(err.Error())
		return
	}

	io.Println("[..] Please push the button on your Hue Bridge")
	for {
		key, retry, err := authenticator.Authenticate()
		if err != nil && retry {
			// this is an expected error, we just wait for the user to push the button
			io.Printf(".")
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil && !retry {
			// this is unexpected, let's print the error and exit
			io.ErrPrintln(err.Error())
			break
		}
		io.Println("\n[OK] Successfully paired openhue with your Hue Bridge!")
		path, err := saveConfig(ip, key)
		if err != nil {
			io.ErrPrintln("[KO] Unable to save config")
		}
		io.Println("[OK] Configuration saved in file", path)
		break
	}
}

func getBridgeIPAddress(io openhue.IOStreams, o *CmdSetupOptions) (string, error) {

	if len(o.bridge) > 0 {
		log.Infof("Bridge IP address set from flag --bridge with value %s", o.bridge)
		io.Printf("[OK] Bridge IP is '%s'\n", o.bridge)
		return o.bridge, nil
	}

	log.Info("Bridge IP address no set from flag --bridge, start lookup via mDNS service discovery")

	b, err := oh.NewBridgeDiscovery(oh.WithTimeout(2 * time.Second)).Discover()
	if err != nil {
		return "", fmt.Errorf("[KO] Unable to discover your Hue Bridge on your local network, you can also visit %s\n", hueBridgeDiscover)
	}

	io.Printf("[OK] Found Hue Bridge with IP '%s'\n", b.IpAddress)
	return b.IpAddress, nil
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
