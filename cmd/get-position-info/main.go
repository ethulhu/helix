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
	"github.com/ethulhu/helix/upnpav/avtransport"
)

var (
	server = flag.String("server", "", "name of server to list")

	ifaceName = flag.String("interface", "", "network interface to discover on (optional)")
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

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	devices, _, err := upnp.DiscoverDevices(ctx, avtransport.Version1, iface)
	if err != nil {
		log.Fatalf("could not discover AVTransport clients: %v", err)
	}

	var transport avtransport.Interface
	for _, device := range devices {
		if client, ok := device.SOAPInterface(avtransport.Version1); ok && device.UDN == *server {
			transport = avtransport.NewClient(client)
			break
		}
	}
	if transport == nil {
		log.Fatalf("could not find AVTransport server %v", *server)
	}

	ctx, _ = context.WithTimeout(context.Background(), 1*time.Second)
	uri, metadata, duration, reltime, err := transport.PositionInfo(ctx)
	if err != nil {
		log.Fatalf("could not get media info: %v", err)
	}
	fmt.Println(uri)
	fmt.Printf("%+v\n", metadata)
	fmt.Println(duration)
	fmt.Println(reltime)
}
