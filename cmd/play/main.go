package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/ethulhu/helix/upnpav/avtransport"
)

var (
	server = flag.String("server", "", "name of server to list")

	uri = flag.String("uri", "", "uri to play")
)

func main() {
	flag.Parse()

	if *server == "" || *uri == "" {
		log.Fatal("must set -server and -uri")
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	clients, _, err := avtransport.Discover(ctx)
	if err != nil {
		log.Fatalf("could not discover AVTransport clients: %v", err)
	}

	var client avtransport.Client
	for _, c := range clients {
		if c.Name() == *server {
			client = c
			break
		}
	}
	if client == nil {
		log.Fatalf("could not find AVTransport server %v", *server)
	}

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	if err := client.Stop(ctx); err != nil {
		log.Fatal(err)
	}

	if err := client.SetCurrentURI(ctx, *uri, nil); err != nil {
		log.Fatal(err)
	}

	if err := client.Play(ctx); err != nil {
		log.Fatal(err)
	}
}
