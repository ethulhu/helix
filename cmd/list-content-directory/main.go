package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

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
	clients, _, err := contentdirectory.Discover(ctx)
	if err != nil {
		log.Fatalf("could not discover ContentDirectory clients: %v", err)
	}

	var client contentdirectory.Client
	for _, c := range clients {
		if c.Name() == *server {
			client = c
			break
		}
	}
	if client == nil {
		log.Fatalf("could not find ContentDirectory server %v", *server)
	}

	ctx, _ = context.WithTimeout(context.Background(), 1*time.Second)
	collections, items, err := client.Browse(ctx, upnpav.Object(*object))
	if err != nil {
		log.Fatalf("could not list ContentDirectory root: %v", err)
	}

	for _, collection := range collections {
		fmt.Printf("%+v\n", collection)
	}
	for _, item := range items {
		fmt.Printf("%+v\n", item)
	}
}
