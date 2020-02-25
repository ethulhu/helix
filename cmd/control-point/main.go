package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/gorilla/mux"
)

var (
	port   = flag.Uint("port", 0, "port to listen on")
	socket = flag.String("socket", "", "path to socket to listen to")
)

var (
	devices     = map[string]*ssdp.Device{}
	devicesLock = sync.Mutex{}
)

func updateDevices() {
	devicesLock.Lock()
	defer devicesLock.Unlock()

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	newDevices, _, err := ssdp.Discover(ctx, ssdp.All)
	if err != nil {
		log.Printf("could not find UPnP devices: %v", err)
		return
	}
	for _, device := range newDevices {
		devices[device.UDN] = device
	}
}

func main() {
	flag.Parse()

	if (*port == 0) == (*socket == "") {
		log.Fatal("must set -socket XOR -port")
	}
	var conn net.Listener
	var err error
	if *port != 0 {
		conn, err = net.Listen("tcp", fmt.Sprintf(":%v", *port))
	} else {
		_ = os.Remove(*socket)
		conn, err = net.Listen("unix", *socket)
		_ = os.Chmod(*socket, 0660)
	}
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer conn.Close()

	updateDevices()
	go func() {
		for _ = range time.Tick(1 * time.Minute) {
			updateDevices()
		}
	}()

	m := mux.NewRouter()
	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("not found: %s %s %s", r.Method, r.URL, r.Form), http.StatusNotFound)
	})

	m.Path("/").
		Methods("GET").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/browse", http.StatusFound)
		})

	m.Path("/browse").
		Methods("GET").
		HandlerFunc(getDirectories)

	m.Path("/browse/{udn}").
		Methods("GET").
		HandlerFunc(getDirectory)

	m.Path("/browse/{udn}/{object}").
		Methods("GET").
		HandlerFunc(needsDirectory("udn", getObject))

	m.Path("/renderer/{udn}").
		Methods("POST").
		MatcherFunc(FormValues("action", "stop")).
		HandlerFunc(needsTransport("udn", stop))

	m.Path("/renderer/{udn}").
		Methods("POST").
		MatcherFunc(FormValues("action", "pause")).
		HandlerFunc(needsTransport("udn", pause))

	m.Path("/renderer/{udn}").
		Methods("POST").
		MatcherFunc(FormValues(
			"action", "play",
			"directory", "{directory}",
			"object", "{object}",
		)).
		HandlerFunc(needsTransport("udn", needsDirectory("directory", playObject)))

	m.Path("/renderer/{udn}").
		Methods("POST").
		MatcherFunc(FormValues("action", "play")).
		HandlerFunc(needsTransport("udn", play))

	log.Printf("starting HTTP server on %v", conn.Addr())
	if err := http.Serve(conn, m); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
