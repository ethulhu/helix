package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ethulhu/helix/cmd/helix-control-point/internal"
	"github.com/gorilla/mux"
)

var (
	port   = flag.Uint("port", 0, "port to listen on")
	socket = flag.String("socket", "", "path to socket to listen to")

	upnpRefresh = flag.Duration("upnp-refresh", 30*time.Second, "how frequently to refresh the UPnP devices")
)

var (
	devices *internal.Devices
	queue   *internal.Queue
)

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

	devices = internal.NewDevices(*upnpRefresh)
	queue = internal.NewQueue()

	m := mux.NewRouter()
	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("not found: %s %s %s", r.Method, r.URL, r.Form), http.StatusNotFound)
	})

	m.Path("/").
		Methods("GET").
		HandlerFunc(getIndexHTML)

	m.Path("/browse").
		Methods("GET").
		HandlerFunc(getDirectories)

	m.Path("/browse/{udn}").
		Methods("GET").
		HandlerFunc(getDirectory)

	m.Path("/browse/{udn}/{object}").
		Methods("GET").
		HandlerFunc(getObject)

	m.Path("/queue").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getQueueJSON)
	// m.Path("/queue").
	// Methods("GET").
	// HandlerFunc(getQueueHTML)

	m.Path("/queue").
		Methods("POST").
		MatcherFunc(FormValues("transport", "{transport}")).
		HandlerFunc(setQueueTransport)
	m.Path("/queue").
		Methods("POST").
		MatcherFunc(FormValues("action", "play")).
		HandlerFunc(playQueue)
	m.Path("/queue").
		Methods("POST").
		MatcherFunc(FormValues("action", "pause")).
		HandlerFunc(pauseQueue)
	m.Path("/queue").
		Methods("POST").
		MatcherFunc(FormValues("action", "stop")).
		HandlerFunc(stopQueue)
	m.Path("/queue").
		Methods("POST").
		MatcherFunc(FormValues(
			"action", "add",
			"position", "last",
			"directory", "{directory}",
			"object", "{object}",
		)).
		HandlerFunc(addObjectToQueue)
	m.Path("/queue").
		Methods("POST").
		MatcherFunc(FormValues(
			"action", "remove",
			"position", "all",
		)).
		HandlerFunc(removeAllFromQueue)

	m.Path("/renderer/{udn}").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getTransportJSON)
	m.Path("/renderer/{udn}").
		Methods("GET").
		HandlerFunc(getTransportHTML)

	m.Path("/renderer/{udn}").
		Methods("POST").
		MatcherFunc(FormValues("action", "stop")).
		HandlerFunc(stop)

	m.Path("/renderer/{udn}").
		Methods("POST").
		MatcherFunc(FormValues("action", "pause")).
		HandlerFunc(pause)

	m.Path("/renderer/{udn}").
		Methods("POST").
		MatcherFunc(FormValues(
			"action", "play",
			"directory", "{directory}",
			"object", "{object}",
		)).
		HandlerFunc(playObject)

	m.Path("/renderer/{udn}").
		Methods("POST").
		MatcherFunc(FormValues("action", "play")).
		HandlerFunc(play)

	log.Printf("starting HTTP server on %v", conn.Addr())
	if err := http.Serve(conn, m); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
