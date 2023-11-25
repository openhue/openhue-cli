package setup

import (
	"context"
	"fmt"
	"github.com/brutella/dnssd"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const serviceName = "_hue._tcp"
const domain = ".local"

// NewCmdDiscover represents the discover command
func NewCmdDiscover() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "discover",
		GroupID: "config",
		Short:   "Hue Bridge discovery",
		Long:    `Discover your Hue Bridge on your local network using the mDNS Service Discovery`,
		Run: func(cmd *cobra.Command, args []string) {

			service := fmt.Sprintf("%s.%s.", strings.Trim(serviceName, "."), strings.Trim(domain, "."))

			foundFn := func(e dnssd.BrowseEntry) {

				for _, ip := range e.IPs {
					if ip.To4() != nil { // we want to display IPv4 address only
						c := color.New(color.FgGreen).Add(color.Bold)
						c.Printf("\nBridge '%s' found with IP '%s'\n", strings.Replace(e.Name, "\\", "", 3), ip)
						os.Exit(0)
					}
				}
			}

			err := dnssd.LookupType(context.Background(), service, foundFn, nil)
			cobra.CheckErr(err)
		},
	}

	return cmd

}
