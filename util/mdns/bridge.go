package mdns

import (
	"context"
	"fmt"
	"github.com/brutella/dnssd"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
	"time"
)

const serviceName = "_hue._tcp"
const domain = ".local"

func DiscoverBridge(result chan *net.IP, timeout time.Duration) {
	// build the service identifier
	service := fmt.Sprintf("%s.%s.", strings.Trim(serviceName, "."), strings.Trim(domain, "."))
	log.Infof("DNS-SD service '%s'", service)
	found := false
	foundFn := func(e dnssd.BrowseEntry) {

		for _, ip := range e.IPs {
			if ip.To4() != nil { // we want to display IPv4 address only
				log.Info("Found bridge with IP ", ip)
				result <- &ip
				found = true
				return
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := dnssd.LookupType(ctx, service, foundFn, nil)
	if err != nil && !found {
		log.Errorf("Unable to lookup bridge with service '%s'. Check error details below:", service)
		log.Error(err)
	}
	<-ctx.Done()
}
