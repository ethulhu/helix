// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

// Binary list-devices lists UPnP devices on the local network.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
)

var (
	timeout = flag.Duration("timeout", 2*time.Second, "how long to wait")

	ifaceName = flag.String("interface", "", "network interface to discover on (optional)")
)

func main() {
	flag.Parse()

	var iface *net.Interface
	if *ifaceName != "" {
		var err error
		iface, err = net.InterfaceByName(*ifaceName)
		if err != nil {
			log.Fatalf("could not find interface %s: %v", *ifaceName, err)
		}
	}

	ctx, _ := context.WithTimeout(context.Background(), *timeout)
	devices, errs, err := ssdp.Discover(ctx, ssdp.All, iface)
	if err != nil {
		log.Fatalf("could not discover URLs: %v", err)
	}
	for _, err := range errs {
		log.Printf("could not get URL: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, device := range devices {
		urns := device.Services()

		sort.Slice(urns, func(i, j int) bool { return urns[i] < urns[j] })
		for _, urn := range urns {
			fmt.Fprintf(w, "%v\t%v\t%v\n", device.Name, device.UDN, urn)
		}
	}
	w.Flush()
}
