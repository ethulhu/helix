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
	"sort"
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

	m.Path("/{udn}").
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
	devices := directories.Devices()
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Name < devices[j].Name
	})

	var deviceLIs []string
	for _, device := range devices {
		deviceLIs = append(deviceLIs, fmt.Sprintf(`<li><a href='/%s'>%s</a></li>`, device.UDN, device.Name))
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Helix Player</title></head>
<body><ul>%s</ul></body>
</html>`, strings.Join(deviceLIs, ""))
}

func getDirectoryHTML(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]

	if _, ok := directories.DeviceByUDN(udn); !ok {
		http.Error(w, fmt.Sprintf("unknown ContentDirectory: %v", udn), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", udn, contentdirectory.Root), http.StatusFound)
}

func getObjectHTML(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]
	object := mux.Vars(r)["object"]

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

	self, err := directory.BrowseMetadata(ctx, upnpav.Object(object))
	if err != nil {
		http.Error(w, fmt.Sprintf("could not fetch object metadata: %v", err), http.StatusInternalServerError)
		return
	}

	switch {
	case len(self.Containers) > 0:
		children, err := directory.BrowseChildren(ctx, upnpav.Object(object))
		if err != nil {
			http.Error(w, fmt.Sprintf("could not fetch object children: %v", err), http.StatusInternalServerError)
			return
		}
		var childrenLIs []string
		for _, container := range children.Containers {
			childrenLIs = append(childrenLIs, fmt.Sprintf(`<li><a href='/%s/%s'>%s</a></li>`, udn, container.ID, container.Title))
		}
		for _, item := range children.Items {
			childrenLIs = append(childrenLIs, fmt.Sprintf(`<li><a href='/%s/%s'>%s</a></li>`, udn, item.ID, item.Title))
		}
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Helix Player</title></head>
<body><ul>%s</ul></body>
</html>`, strings.Join(childrenLIs, ""))
	case len(self.Items) > 0:
		item := self.Items[0]

		// TODO: switch by itemClass, e.g. VideoItem, AudioItem.
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Helix Player</title></head>
<body><audio src='/%s/%s?accept=audio/*' controls></audio></body>
</html>`, udn, item.ID)
	default:
		// I think this is impossible, but I can't be sure.
		http.Error(w, "object is neither item nor container?!", http.StatusInternalServerError)
	}
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

	proxyGet(w, uri)
}

func proxyGet(w http.ResponseWriter, uri string) {
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
	w.WriteHeader(rsp.StatusCode)

	io.Copy(w, rsp.Body)
}
