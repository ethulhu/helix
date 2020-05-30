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
	"time"

	"github.com/ethulhu/helix/upnp"
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
	urls, errs, err := upnp.DiscoverURLs(ctx, upnp.All, iface)
	if err != nil {
		log.Fatalf("could not discover URLs: %v", err)
	}
	for _, err := range errs {
		log.Printf("could not get URL: %v", err)
	}

	for _, url := range urls {
		fmt.Println(url)
	}
}
