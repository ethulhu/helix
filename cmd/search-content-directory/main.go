// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/ethulhu/helix/flag"
	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
)

var (
	query = flag.Custom("query", "", "query to run", func(raw string) (interface{}, error) {
		return search.Parse(raw)
	})

	object = flag.String("object", "0", "object to list (0 means root)")
	server = flag.String("server", "", "UDN of server to list")

	timeout = flag.Duration("timeout", 2*time.Second, "how long to wait for device discovery")
	iface   = flag.Custom("interface", "", "network interface to discover on (optional)", func(raw string) (interface{}, error) {
		if raw == "" {
			return (*net.Interface)(nil), nil
		}
		return net.InterfaceByName(raw)
	})
)

func main() {
	flag.Parse()

	if *server == "" {
		log.Fatal("must set -server")
	}
	query := (*query).(search.Criteria)
	iface := (*iface).(*net.Interface)

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

	didl, err := directory.Search(ctx, upnpav.ObjectID(*object), query)
	if err != nil {
		log.Printf("could not search ContentDirectory: %v", err)

		caps, err := directory.SearchCapabilities(ctx)
		if err != nil {
			log.Fatalf("could not get Search Capabilities: %v", err)
		}
		log.Printf("ContentDirectory supports: %v", caps)
		os.Exit(1)
	}

	for _, collection := range didl.Containers {
		fmt.Printf("%+v\n", collection)
	}
	for _, item := range didl.Items {
		fmt.Printf("%+v\n", item)
	}
}
