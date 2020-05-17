//go:generate go run aletheia.icu/broccoli -src=static -o assets -var=assets
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
	"github.com/ethulhu/helix/upnpav/controlpoint"
	"github.com/gorilla/mux"
)

var (
	port   = flag.Uint("port", 0, "port to listen on")
	socket = flag.String("socket", "", "path to socket to listen to")

	debugAssetsPath = flag.String("debug-assets-path", "", "path to assets to load from filesystem, for development")

	ifaceName      = flag.String("interface", "", "network interface to discover on (optional)")
	initialRefresh = flag.Duration("initial-upnp-refresh", 5*time.Second, "how frequently discover new UPnP devices when the server hasn't found any yet")
	stableRefresh  = flag.Duration("stable-upnp-refresh", 30*time.Second, "how frequently discover new UPnP devices when the server has found some already")
)

var (
	directories *upnp.DeviceCache
	transports  *upnp.DeviceCache

	controlLoop = controlpoint.NewLoop()
	trackList   = controlpoint.NewTrackList()
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

	deviceCacheOptions := upnp.DeviceCacheOptions{
		InitialRefresh: *initialRefresh,
		StableRefresh:  *stableRefresh,
		Interface:      iface,
	}
	directories = upnp.NewDeviceCache(contentdirectory.Version1, deviceCacheOptions)
	transports = upnp.NewDeviceCache(avtransport.Version1, deviceCacheOptions)

	// TODO: support multiple Queues.
	controlLoop.SetQueue(trackList)

	m := mux.NewRouter()
	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("not found: %v %v %v", r.Method, r.URL, r.Form)
		if r.URL.Path != "/favicon.ico" {
			log.Print(msg)
		}
		http.Error(w, msg, http.StatusNotFound)
	})

	// ContentDirectory routes.

	m.Path("/directories/").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getDirectoriesJSON)

	m.Path("/directories/{udn}").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getDirectoryJSON)

	m.Path("/directories/{udn}/{object}").
		Methods("GET", "HEAD").
		HeadersRegexp("Accept", "(application|text)/json").
		Queries("search", "{query}").
		HandlerFunc(searchUnderObjectJSON)

	m.Path("/directories/{udn}/{object}").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getObjectJSON)

	m.Path("/directories/{udn}/{object}").
		Methods("GET", "HEAD").
		Queries("accept", "{mimetype}").
		HandlerFunc(getObjectByType)

	// AVTransport routes.

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

	// Control Point routes.

	m.Path("/control-point/").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getControlPointJSON)

	m.Path("/control-point/").
		Methods("POST").
		MatcherFunc(httputil.FormValues("transport", "{udn}")).
		HandlerFunc(setControlPointTransport)

	m.Path("/control-point/").
		Methods("POST").
		MatcherFunc(httputil.FormValues("state", "playing")).
		HandlerFunc(playControlPoint)

	m.Path("/control-point/").
		Methods("POST").
		MatcherFunc(httputil.FormValues("state", "paused")).
		HandlerFunc(pauseControlPoint)

	m.Path("/control-point/").
		Methods("POST").
		MatcherFunc(httputil.FormValues("state", "stopped")).
		HandlerFunc(stopControlPoint)

	m.Path("/control-point/").
		Methods("POST").
		MatcherFunc(httputil.FormValues("state", "{unknown}")).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			state := mux.Vars(r)["unknown"]
			http.Error(w, fmt.Sprintf("unknown state: %v", state), http.StatusBadRequest)
		})

	// Queue routes.

	m.Path("/queue/").
		Methods("GET").
		HeadersRegexp("Accept", "(application|text)/json").
		HandlerFunc(getQueueJSON)

	// TODO: support multiple queues.
	m.Path("/queue/").
		Methods("POST").
		MatcherFunc(httputil.FormValues(
			"directory", "{udn}",
			"object", "{object}",
		)).
		HandlerFunc(appendToQueue)

	m.Path("/queue/").
		Methods("POST").
		MatcherFunc(httputil.FormValues(
			"directory", "{udn}",
			"object", "{object}",
			"position", "{position}",
		)).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not implemented", http.StatusNotImplemented)
		})

	m.Path("/queue/").
		Methods("POST").
		MatcherFunc(httputil.FormValues(
			"current", "{id}",
		)).
		HandlerFunc(setCurrentQueueTrack)

	m.Path("/queue/").
		Methods("POST").
		MatcherFunc(httputil.FormValues(
			"remove", "all",
		)).
		HandlerFunc(removeAllFromQueue)

	m.Path("/queue/").
		Methods("POST").
		MatcherFunc(httputil.FormValues(
			"remove", "{id}",
		)).
		HandlerFunc(removeTrackFromQueue)

	// Assets routes.

	if *debugAssetsPath != "" {
		m.PathPrefix("/").
			Methods("GET").
			Handler(http.FileServer(httputil.TryFiles{http.Dir(*debugAssetsPath)}))
	} else {
		m.PathPrefix("/").
			Methods("GET").
			Handler(http.FileServer(httputil.TryFiles{httputil.BroccoliFS{"static", assets}}))
	}

	log.Printf("starting HTTP server on %v", conn.Addr())
	if err := http.Serve(conn, m); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
