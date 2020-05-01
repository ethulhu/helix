//go:generate broccoli -src=static -o assets -var=assets
package main

import (
	"encoding/json"
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

	debugAssetsPath = flag.String("debug-assets-path", "", "path to assets to load from filesystem, for development")

	ifaceName   = flag.String("interface", "", "network interface to discover on (optional)")
	upnpRefresh = flag.Duration("upnp-refresh", 30*time.Second, "how frequently to refresh the UPnP devices")
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

func getDirectoriesJSON(w http.ResponseWriter, r *http.Request) {
	devices := directories.Devices()

	data := []directory{}
	for _, device := range devices {
		data = append(data, directoryFromDevice(device))
	}

	blob, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("could not marshal JSON: %v", err))
	}
	w.Write(blob)
}

func getDirectoryJSON(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]

	device, ok := directories.DeviceByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown ContentDirectory: %v", udn), http.StatusNotFound)
		return
	}

	data := directoryFromDevice(device)

	blob, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("could not marshal JSON: %v", err))
	}
	w.Write(blob)
}

func getObjectJSON(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]
	objectID := mux.Vars(r)["object"]

	device, ok := directories.DeviceByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown ContentDirectory: %s", udn), http.StatusNotFound)
		return
	}
	soapClient, ok := device.Client(contentdirectory.Version1)
	if !ok {
		http.Error(w, fmt.Sprintf("UPnP device exists but is not a ContentDirectory: %s", udn), http.StatusInternalServerError)
		return
	}
	directory := contentdirectory.NewClient(soapClient)

	ctx := r.Context()
	self, err := directory.BrowseMetadata(ctx, upnpav.Object(objectID))
	if err != nil {
		http.Error(w, fmt.Sprintf("could not fetch object metadata: %v", err), http.StatusInternalServerError)
		return
	}

	data := object{}
	switch {
	case len(self.Containers) == 1 && len(self.Items) == 0:
		data = objectFromContainer(udn, self.Containers[0])

		children, err := directory.BrowseChildren(ctx, upnpav.Object(objectID))
		if err != nil {
			http.Error(w, fmt.Sprintf("could not fetch object children: %v", err), http.StatusInternalServerError)
			return
		}
		for _, container := range children.Containers {
			data.Children = append(data.Children, objectFromContainer(udn, container))
		}
		for _, item := range children.Items {
			data.Children = append(data.Children, objectFromItem(udn, item))
		}

	case len(self.Containers) == 0 && len(self.Items) == 1:
		data = objectFromItem(udn, self.Items[0])

	default:
		http.Error(w, fmt.Sprintf("object has %v containers and %v items", len(self.Containers), len(self.Items)), http.StatusInternalServerError)
		return
	}

	blob, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("could not marshal JSON: %v", err))
	}
	w.Write(blob)
}

func getObjectByType(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]
	object := mux.Vars(r)["object"]
	mimetypeRaw := mux.Vars(r)["mimetype"]

	log.Printf("%v udn %q object %q MIME-type %q", r.Method, udn, object, mimetypeRaw)

	mimetype, _, err := mime.ParseMediaType(mimetypeRaw)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid MIME-Type %q: %v", mimetypeRaw, err), http.StatusBadRequest)
		return
	}

	device, ok := directories.DeviceByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown ContentDirectory: %s", udn), http.StatusNotFound)
		return
	}
	soapClient, ok := device.Client(contentdirectory.Version1)
	if !ok {
		http.Error(w, fmt.Sprintf("UPnP device exists but is not a ContentDirectory: %s", udn), http.StatusInternalServerError)
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
	mimeParts := strings.Split(mimetype, "/")
	for _, r := range item.Resources {
		if r.ProtocolInfo.Protocol != upnpav.ProtocolHTTP {
			continue
		}

		if strings.HasPrefix(r.ProtocolInfo.ContentFormat, mimetype) {
			uri = r.URI
			break
		}

		if mimeParts[1] == "*" && strings.HasPrefix(r.ProtocolInfo.ContentFormat, mimeParts[0]+"/") {
			uri = r.URI
			break
		}
	}

	if uri == "" {
		http.Error(w, fmt.Sprintf("could not find matching resource for MIME-type %q", mimetype), http.StatusNotFound)
		return
	}

	proxyDo(w, r.Method, uri)
}

func proxyDo(w http.ResponseWriter, method, uri string) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rsp, err := http.DefaultClient.Do(req)
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
	w.WriteHeader(rsp.StatusCode)

	io.Copy(w, rsp.Body)
}
