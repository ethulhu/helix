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
	"text/template"
	"time"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
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
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			discoverCtx, _ := context.WithTimeout(ctx, 2*time.Second)
			devices, _, err := ssdp.Discover(discoverCtx, contentdirectory.Version1)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := directoriesTmpl.Execute(w, devices); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})

	m.Path("/browse/{udn}").
		Methods("GET").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			udn := mustVar(r, "udn")
			http.Redirect(w, r, fmt.Sprintf("/browse/%v/0", udn), http.StatusFound)
		})

	m.Path("/browse/{udn}/{object}").
		Methods("GET").
		HandlerFunc(needsDirectory("udn", func(w http.ResponseWriter, r *http.Request) {
			object := mustVar(r, "object")
			udn := mustVar(r, "udn")

			ctx := r.Context()
			directory := ctx.Value("ContentDirectory").(contentdirectory.Client)

			didl, err := directory.Browse(ctx, contentdirectory.BrowseChildren, upnpav.Object(object))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			args := struct {
				DIDL *upnpav.DIDL
				UDN  string
			}{didl, udn}
			if err := browseTmpl.Execute(w, args); err != nil {
				log.Printf("error rendering %v: %v", r.URL.Path, err)
				return
			}
		}))

	m.Path("/renderer/{udn}").
		Methods("POST").
		Queries("action", "stop").
		HandlerFunc(needsTransport("udn", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			transport := ctx.Value("AVTransport").(avtransport.Client)
			if err := transport.Stop(ctx); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			maybeRedirect(w, r)
		}))

	m.Path("/renderer/{udn}").
		Methods("POST").
		Queries("action", "pause").
		HandlerFunc(needsTransport("udn", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			transport := ctx.Value("AVTransport").(avtransport.Client)
			if err := transport.Pause(ctx); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			maybeRedirect(w, r)
		}))

	m.Path("/renderer/{udn}").
		Methods("POST").
		Queries(
			"action", "play",
			"directory", "{directory}",
			"object", "{object}",
		).
		HandlerFunc(needsTransport("udn", needsDirectory("directory",
			func(w http.ResponseWriter, r *http.Request) {
				object := mustVar(r, "object")

				ctx := r.Context()
				transport := ctx.Value("AVTransport").(avtransport.Client)
				directory := ctx.Value("ContentDirectory").(contentdirectory.Client)

				didl, err := directory.Browse(ctx, contentdirectory.BrowseMetadata, upnpav.Object(object))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if didl == nil {
					http.Error(w, fmt.Sprintf("could not find object %s", object), http.StatusNotFound)
					return
				}

				if err := transport.Stop(ctx); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if err := transport.SetCurrentURI(ctx, didl.Items[0].Resources[0].URI, didl); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if err := transport.Play(ctx); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				maybeRedirect(w, r)
			})))

	m.Path("/renderer/{udn}").
		Methods("POST").
		Queries("action", "play").
		HandlerFunc(needsTransport("udn", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			transport := ctx.Value("AVTransport").(avtransport.Client)
			if err := transport.Play(ctx); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			maybeRedirect(w, r)
		}))

	log.Printf("starting HTTP server on %v", conn.Addr())
	if err := http.Serve(conn, m); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

var (
	browseTmpl = template.Must(template.New("/browse").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
</head>
<body>
	{{ $udn := .UDN }}
	<ul>
	{{ range $index, $container := .DIDL.Containers }}
		<li><a href='/browse/{{ $udn }}/{{ $container.ID }}'>{{ $container.Title }}</a></li>
	{{ end }}
	</ul>
	<ul>
	{{ range $index, $item := .DIDL.Items }}
		<li><a href='/browse/{{ $udn }}/{{ $item.ID }}'>{{ $item.Title }}</a></li>
	{{ end }}
	</ul>
</body>
</html>`))

	directoriesTmpl = template.Must(template.New("/browse").Parse(`<!DOCTYPE html>
<html lang='en'>
<head>
	<meta charset='utf-8'>
	<meta name='viewport' content='width=device-width, initial-scale=1.0'>
</head>
<body>
	<ul>
	{{ range $index, $device := . }}
		<li><a href='/browse/{{ $device.UDN }}'>{{ $device.Name }}</a></li>
	{{ end }}
	</ul>
</body>
</html>`))
)

func mustVar(r *http.Request, key string) string {
	value, ok := mux.Vars(r)[key]
	if !ok {
		panic(fmt.Sprintf("gorilla/mux did not provide parameter %q", key))
	}
	return value
}

func maybeRedirect(w http.ResponseWriter, r *http.Request) {
	if redirect := r.Form.Get("redirect"); redirect != "" {
		http.Redirect(w, r, redirect, http.StatusFound)
	}
}

func needsDirectory(key string, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		udn := mustVar(r, key)
		ctx := r.Context()

		devicesLock.Lock()
		device, ok := devices[udn]
		devicesLock.Unlock()

		if !ok {
			http.Error(w, fmt.Sprintf("could not find ContentDirectory %s", udn), http.StatusNotFound)
			return
		}

		soapClient, ok := device.Client(contentdirectory.Version1)
		if !ok {
			http.Error(w, fmt.Sprintf("found a device %s, but it was not an ContentDirectory", udn), http.StatusNotFound)
			return
		}
		directory := contentdirectory.NewClient(soapClient)

		newCtx := context.WithValue(ctx, "ContentDirectory", directory)
		f(w, r.WithContext(newCtx))
	}
}
func needsTransport(key string, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		udn := mustVar(r, key)
		ctx := r.Context()

		devicesLock.Lock()
		device, ok := devices[udn]
		devicesLock.Unlock()

		if !ok {
			http.Error(w, fmt.Sprintf("could not find AVTransport %s", udn), http.StatusNotFound)
			return
		}

		soapClient, ok := device.Client(avtransport.Version1)
		if !ok {
			http.Error(w, fmt.Sprintf("found a device %s, but it was not an AVTransport", udn), http.StatusNotFound)
			return
		}
		transport := avtransport.NewClient(soapClient)

		newCtx := context.WithValue(ctx, "AVTransport", transport)
		f(w, r.WithContext(newCtx))
	}
}
