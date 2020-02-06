package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

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

	ctx, _ = context.WithTimeout(context.Background(), 1*time.Second)
	state, err := client.TransportInfo(ctx)
	if err != nil {
		log.Fatalf("could not get media info: %v", err)
	}
	fmt.Println(state)
}
