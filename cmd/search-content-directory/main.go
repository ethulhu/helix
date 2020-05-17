package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
)

var (
	query  = flag.String("query", "", "query to run")
	object = flag.String("object", "0", "object to list (0 means root)")
	server = flag.String("server", "", "name of server to list")

	ifaceName = flag.String("interface", "", "network interface to discover on (optional)")
	timeout   = flag.Duration("timeout", 2*time.Second, "how long to wait for device discovery")
)

func main() {
	flag.Parse()

	if *server == "" || *query == "" {
		log.Fatal("must set -server and -query")
	}

	criteria, err := search.Parse(*query)
	if err != nil {
		log.Fatalf("could not parse query %q: %v", *query, err)
	}

	var iface *net.Interface
	if *ifaceName != "" {
		var err error
		iface, err = net.InterfaceByName(*ifaceName)
		if err != nil {
			log.Fatalf("could not find interface %s: %v", *ifaceName, err)
		}
	}

	ctx := context.Background()

	opts := upnp.DeviceCacheOptions{
		InitialRefresh: *timeout,
		StableRefresh:  *timeout,
		Interface:      iface,
	}
	directories := upnp.NewDeviceCache(contentdirectory.Version1, opts)

	var directory contentdirectory.Client
	for {
		time.Sleep(*timeout)
		if device, ok := directories.DeviceByUDN(*server); ok {
			client, ok := device.SOAPClient(contentdirectory.Version1)
			if !ok {
				log.Fatal("device exists, but has no ContentDirectory service")
			}
			directory = contentdirectory.NewClient(client)
			break
		}
		log.Print("could not find ContentDirectory; sleeping and retrying")
	}

	didl, err := directory.Search(ctx, upnpav.ObjectID(*object), criteria)
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
