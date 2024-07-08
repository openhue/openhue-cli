package setup

import (
	oh "github.com/openhue/openhue-go"
	"github.com/spf13/cobra"
	"openhue-cli/openhue"
	"time"
)

// NewCmdDiscover represents the discover command
func NewCmdDiscover(io openhue.IOStreams) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "discover",
		GroupID: "config",
		Short:   "Hue Bridge discovery",
		Long:    `Discover your Hue Bridge on your local network using the mDNS Service Discovery`,
		Run: func(cmd *cobra.Command, args []string) {

			b, err := oh.NewBridgeDiscovery(oh.WithTimeout(2 * time.Second)).Discover()
			cobra.CheckErr(err)
			io.Printf("%s\n", b.IpAddress)
		},
	}

	return cmd
}
