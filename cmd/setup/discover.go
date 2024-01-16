package setup

import (
	"github.com/spf13/cobra"
	"net"
	"openhue-cli/openhue"
	"openhue-cli/util/mdns"
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

			ip := make(chan *net.IP)
			go mdns.DiscoverBridge(ip, 5*time.Second)
			io.Printf("%s\n", <-ip)
		},
	}

	return cmd
}
