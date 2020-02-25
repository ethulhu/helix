package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

var (
	object = flag.String("object", "0", "object to list (0 means root)")
	server = flag.String("server", "", "name of server to list")
)

func main() {
	flag.Parse()

	if *server == "" {
		log.Fatal("must set -server")
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	devices, _, err := ssdp.Discover(ctx, contentdirectory.Version1)
	if err != nil {
		log.Fatalf("could not discover ContentDirectory clients: %v", err)
	}

	var client contentdirectory.Client
	for _, device := range devices {
		if soapClient, ok := device.Client(contentdirectory.Version1); ok && device.Name == *server {
			client = contentdirectory.NewClient(soapClient)
			break
		}
	}
	if client == nil {
		log.Fatalf("could not find ContentDirectory server %v", *server)
	}

	ctx, _ = context.WithTimeout(context.Background(), 1*time.Second)
	didl, err := client.Browse(ctx, contentdirectory.BrowseChildren, upnpav.Object(*object))
	if err != nil {
		log.Fatalf("could not list ContentDirectory root: %v", err)
	}

	for _, collection := range didl.Containers {
		fmt.Printf("%+v\n", collection)
	}
	for _, item := range didl.Items {
		fmt.Printf("%+v\n", item)
	}
}
