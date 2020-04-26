package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/gorilla/mux"
)

var (
	port   = flag.Uint("port", 0, "port to listen on")
	socket = flag.String("socket", "", "path to socket to listen to")

	upnpRefresh = flag.Duration("upnp-refresh", 30*time.Second, "how frequently to refresh the UPnP devices")

	ifaceName = flag.String("interface", "", "network interface to discover on (optional)")
)

var (
	directories *upnp.DeviceCache
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

	m := mux.NewRouter()
	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("not found: %s %s %s", r.Method, r.URL, r.Form)
		if r.URL.Path != "/favicon.ico" {
			log.Print(msg)
		}
		http.Error(w, msg, http.StatusNotFound)
	})

	m.Path("/").
		Methods("GET").
		HandlerFunc(getIndexHTML)

	m.Path("/{udn}/").
		Methods("GET").
		HandlerFunc(getDirectoryHTML)

	m.Path("/{udn}/{object}").
		Methods("GET").
		Queries("accept", "{mimetype}").
		HandlerFunc(getObjectByType)

	m.Path("/{udn}/{object}").
		Methods("GET").
		HandlerFunc(getObjectHTML)

	log.Printf("starting HTTP server on %v", conn.Addr())
	if err := http.Serve(conn, m); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func getIndexHTML(w http.ResponseWriter, r *http.Request) {
}
func getDirectoryHTML(w http.ResponseWriter, r *http.Request) {
}
func getObjectHTML(w http.ResponseWriter, r *http.Request) {
}
func getObjectByType(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]
	object := mux.Vars(r)["object"]
	mimetypeRaw := mux.Vars(r)["mimetype"]

	log.Printf("GET udn %q object %q MIME-type %q", udn, object, mimetypeRaw)

	mimetype, _, err := mime.ParseMediaType(mimetypeRaw)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid MIME-Type %q: %v", mimetypeRaw, err), http.StatusBadRequest)
		return
	}

	device, ok := directories.DeviceByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find device %q", udn), http.StatusNotFound)
		return
	}
	soapClient, ok := device.Client(contentdirectory.Version1)
	if !ok {
		http.Error(w, fmt.Sprintf("device %q is not a ContentDirectory", udn), http.StatusInternalServerError)
		return
	}
	directory := contentdirectory.NewClient(soapClient)

	// find the object.
	ctx := r.Context()
	self, err := directory.BrowseMetadata(ctx, upnpav.Object(object))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(self.Items) == 0 {
		http.Error(w, fmt.Sprintf("object %q is not an Item", object), http.StatusBadRequest)
		return
	}
	item := self.Items[0]

	uri := ""
	contentType := ""
	mimeParts := strings.Split(mimetype, "/")
	for _, r := range item.Resources {
		if r.ProtocolInfo.Protocol != upnpav.ProtocolHTTP {
			continue
		}

		if strings.HasPrefix(r.ProtocolInfo.ContentFormat, mimetype) {
			uri = r.URI
			contentType = mimetype
			break
		}

		if mimeParts[1] == "*" && strings.HasPrefix(r.ProtocolInfo.ContentFormat, mimeParts[0]+"/") {
			uri = r.URI
			contentType = r.ProtocolInfo.ContentFormat
			break
		}
	}

	if uri == "" {
		http.Error(w, fmt.Sprintf("could not find matching resource for MIME-type %q", mimetype), http.StatusNotFound)
		return
	}

	proxyGet(w, uri, contentType)
}

func proxyGet(w http.ResponseWriter, uri, contentType string) {
	rsp, err := http.Get(uri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rsp.Body.Close()

	for k, vs := range rsp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(rsp.StatusCode)

	io.Copy(w, rsp.Body)
}
