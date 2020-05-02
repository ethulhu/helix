//go:generate broccoli -src=static -o assets -var=assets
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ethulhu/helix/httputil"
	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/gorilla/mux"
)

var (
	port   = flag.Uint("port", 0, "port to listen on")
	socket = flag.String("socket", "", "path to socket to listen to")

	debugAssetsPath = flag.String("debug-assets-path", "", "path to assets to load from filesystem, for development")

	ifaceName   = flag.String("interface", "", "network interface to discover on (optional)")
	upnpRefresh = flag.Duration("upnp-refresh", 30*time.Second, "how frequently to refresh the UPnP devices")
)

var (
	directories *upnp.DeviceCache
	transports  *upnp.DeviceCache
)

func main() {
	flag.Parse()

	var iface *net.Interface
	if *ifaceName != "" {
		var err error
		iface, err = net.InterfaceByName(*ifaceName)
		if err != nil {
			log.Fatalf("could not find interface %s: %v", *ifaceName, err)
		}
	}

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

	directories = upnp.NewDeviceCache(contentdirectory.Version1, *upnpRefresh, iface)
	transports = upnp.NewDeviceCache(avtransport.Version1, *upnpRefresh, iface)

	m := mux.NewRouter()
	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("not found: %v %v %v", r.Method, r.URL, r.Form)
		if r.URL.Path != "/favicon.ico" {
			log.Print(msg)
		}
		http.Error(w, msg, http.StatusNotFound)
	})

	m.Path("/directories/").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getDirectoriesJSON)

	m.Path("/directories/{udn}").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getDirectoryJSON)

	m.Path("/directories/{udn}/{object}").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getObjectJSON)

	m.Path("/directories/{udn}/{object}").
		Methods("GET", "HEAD").
		Queries("accept", "{mimetype}").
		HandlerFunc(getObjectByType)

	m.Path("/transports/").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getTransportsJSON)

	m.Path("/transports/{udn}").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getTransportJSON)

	m.Path("/transports/{udn}").
		Methods("POST").
		MatcherFunc(httputil.FormValues("action", "play")).
		HandlerFunc(playTransport)

	m.Path("/transports/{udn}").
		Methods("POST").
		MatcherFunc(httputil.FormValues("action", "pause")).
		HandlerFunc(pauseTransport)

	m.Path("/transports/{udn}").
		Methods("POST").
		MatcherFunc(httputil.FormValues("action", "stop")).
		HandlerFunc(stopTransport)

	if *debugAssetsPath != "" {
		m.PathPrefix("/").
			Methods("GET").
			Handler(http.FileServer(http.Dir(*debugAssetsPath)))
	} else {
		m.PathPrefix("/").
			Methods("GET").
			Handler(assets.Serve("static"))
	}

	log.Printf("starting HTTP server on %v", conn.Addr())
	if err := http.Serve(conn, m); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
