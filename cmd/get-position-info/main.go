package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav/avtransport"
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
	devices, _, err := ssdp.Discover(ctx, avtransport.Version1)
	if err != nil {
		log.Fatalf("could not discover AVTransport clients: %v", err)
	}

	var client avtransport.Client
	for _, device := range devices {
		if soapClient, ok := device.Client(avtransport.Version1); ok && device.Name == *server {
			client = avtransport.NewClient(soapClient)
			break
		}
	}
	if client == nil {
		log.Fatalf("could not find AVTransport server %v", *server)
	}

	ctx, _ = context.WithTimeout(context.Background(), 1*time.Second)
	uri, metadata, duration, reltime, err := client.PositionInfo(ctx)
	if err != nil {
		log.Fatalf("could not get media info: %v", err)
	}
	fmt.Println(uri)
	fmt.Printf("%+v\n", metadata)
	fmt.Println(duration)
	fmt.Println(reltime)
}