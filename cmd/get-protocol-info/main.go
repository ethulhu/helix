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
	"sort"
	"time"

	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav/connectionmanager"
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
	devices, _, err := upnp.DiscoverDevices(ctx, connectionmanager.Version1, iface)
	if err != nil {
		log.Fatalf("could not discover ConnectionManager clients: %v", err)
	}

	var manager connectionmanager.Interface
	for _, device := range devices {
		if client, ok := device.SOAPInterface(connectionmanager.Version1); ok && device.UDN == *server {
			manager = connectionmanager.NewClient(client)
			break
		}
	}
	if manager == nil {
		log.Fatalf("could not find ConnectionManager server %v", *server)
	}

	ctx, _ = context.WithTimeout(context.Background(), 1*time.Second)
	sources, sinks, err := manager.ProtocolInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("sources:")
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].String() < sources[j].String()
	})
	for _, source := range sources {
		fmt.Println(source)
	}

	fmt.Println()

	fmt.Println("sinks:")
	sort.Slice(sinks, func(i, j int) bool {
		return sinks[i].String() < sinks[j].String()
	})
	for _, sink := range sinks {
		fmt.Println(sink)
	}
}
