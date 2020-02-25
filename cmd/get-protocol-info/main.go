package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav/connectionmanager"
)

var (
	server = flag.String("server", "", "name of server to list")
)

func main() {
	flag.Parse()

	if *server == "" {
		log.Fatal("must set -server")
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	devices, _, err := ssdp.Discover(ctx, connectionmanager.Version1)
	if err != nil {
		log.Fatalf("could not discover ConnectionManager clients: %v", err)
	}

	var client connectionmanager.Client
	for _, device := range devices {
		if soapClient, ok := device.Client(connectionmanager.Version1); ok && device.Name == *server {
			client = connectionmanager.NewClient(soapClient)
			break
		}
	}
	if client == nil {
		log.Fatalf("could not find ConnectionManager server %v", *server)
	}

	ctx, _ = context.WithTimeout(context.Background(), 1*time.Second)
	sources, sinks, err := client.ProtocolInfo(ctx)
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
