// Binary list-devices lists UPnP devices on the local network.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
)

var (
	timeout = flag.Duration("timeout", 2*time.Second, "how long to wait")
)

func main() {
	flag.Parse()

	ctx, _ := context.WithTimeout(context.Background(), *timeout)
	devices, errs, err := ssdp.Discover(ctx, ssdp.All)
	if err != nil {
		log.Fatalf("could not discover URLs: %v", err)
	}
	for _, err := range errs {
		log.Printf("could not get URL: %v", err)
	}

	for _, device := range devices {
		urns := device.Services()

		sort.Slice(urns, func(i, j int) bool { return urns[i] < urns[j] })
		for _, urn := range urns {
			fmt.Printf("%v\t%v\t%v\n", device.Name, device.UDN, urn)
		}
	}
}
