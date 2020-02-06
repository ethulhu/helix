// Binary list-devices lists UPnP devices on the local network.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
)

var (
	timeout = flag.Duration("timeout", 2*time.Second, "how long to wait")
)

func main() {
	flag.Parse()

	ctx, _ := context.WithTimeout(context.Background(), *timeout)
	urls, errs, err := ssdp.DiscoverURLs(ctx, ssdp.All)
	if err != nil {
		log.Fatalf("could not discover URLs: %v", err)
	}
	for _, err := range errs {
		log.Printf("could not get URL: %v", err)
	}

	for _, url := range urls {
		fmt.Println(url)
	}
}
