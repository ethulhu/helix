// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

var (
	server = flag.String("server", "", "name of server to list")

	ifaceName = flag.String("interface", "", "network interface to discover on (optional)")
	timeout   = flag.Duration("timeout", 2*time.Second, "how long to wait for device discovery")
)

func main() {
	flag.Parse()

	if *server == "" {
		log.Fatal("must set -server")
	}

	var iface *net.Interface
	if *ifaceName != "" {
		var err error
		iface, err = net.InterfaceByName(*ifaceName)
		if err != nil {
			log.Fatalf("could not find interface %s: %v", *ifaceName, err)
		}
	}

	ctx := context.Background()

	opts := upnp.DeviceCacheOptions{
		InitialRefresh: *timeout,
		StableRefresh:  *timeout,
		Interface:      iface,
	}
	directories := upnp.NewDeviceCache(contentdirectory.Version1, opts)

	var directory contentdirectory.Interface
	for {
		time.Sleep(*timeout)
		if device, ok := directories.DeviceByUDN(*server); ok {
			client, ok := device.SOAPInterface(contentdirectory.Version1)
			if !ok {
				log.Fatal("device exists, but has no ContentDirectory service")
			}
			directory = contentdirectory.NewClient(client)
			break
		}
		log.Print("could not find ContentDirectory; sleeping and retrying")
	}

	caps, err := directory.SearchCapabilities(ctx)
	if err != nil {
		log.Fatalf("could not get search capabilities: %v", err)
	}
	fmt.Println(caps)
}
