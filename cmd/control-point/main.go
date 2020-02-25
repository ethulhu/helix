package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
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
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			udn := mustVar(r, "udn")
			object := mustVar(r, "object")

			ctx := r.Context()
			discoverCtx, _ := context.WithTimeout(ctx, 2*time.Second)
			devices, _, err := ssdp.Discover(discoverCtx, ssdp.URN(udn))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var directory contentdirectory.Client
			for _, device := range devices {
				if soapClient, ok := device.Client(contentdirectory.Version1); ok && device.UDN == udn {
					directory = contentdirectory.NewClient(soapClient)
				}
			}
			if directory == nil {
				http.Error(w, fmt.Sprintf("could not find ContentDirectory %s", udn), http.StatusNotFound)
				return
			}

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
		})

	m.Path("/renderer/{udn}").
		Methods("POST").
		Queries("action", "stop").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			udn := mustVar(r, "udn")

			ctx := r.Context()
			discoverCtx, _ := context.WithTimeout(ctx, 2*time.Second)
			devices, _, err := ssdp.Discover(discoverCtx, ssdp.URN(udn))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var transport avtransport.Client
			for _, device := range devices {
				if soapClient, ok := device.Client(avtransport.Version1); ok && device.UDN == udn {
					transport = avtransport.NewClient(soapClient)
					break
				}
			}
			if transport == nil {
				http.Error(w, fmt.Sprintf("could not find renderer %s", udn), http.StatusNotFound)
				return
			}
			if err := transport.Stop(ctx); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if redirect := r.Form.Get("redirect"); redirect != "" {
				http.Redirect(w, r, redirect, http.StatusFound)
			}
		})

	m.Path("/renderer/{udn}").
		Methods("POST").
		Queries("action", "pause").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			udn := mustVar(r, "udn")

			ctx := r.Context()
			discoverCtx, _ := context.WithTimeout(ctx, 2*time.Second)
			devices, _, err := ssdp.Discover(discoverCtx, ssdp.URN(udn))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for _, device := range devices {
				if soapClient, ok := device.Client(avtransport.Version1); ok && device.UDN == udn {
					transport := avtransport.NewClient(soapClient)
					if err := transport.Pause(ctx); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
			}
			http.Error(w, fmt.Sprintf("could not find renderer %s", udn), http.StatusNotFound)
		})

	m.Path("/renderer/{udn}").
		Methods("POST").
		Queries(
			"action", "play",
			"directory", "{directory}",
			"object", "{object}",
		).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			directoryUDN := mustVar(r, "directory")
			object := mustVar(r, "object")
			transportUDN := mustVar(r, "udn")

			ctx := r.Context()
			discoverCtx, _ := context.WithTimeout(ctx, 2*time.Second)
			devices, _, err := ssdp.Discover(discoverCtx, ssdp.All)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var directory contentdirectory.Client
			var transport avtransport.Client
			for _, device := range devices {
				if soapClient, ok := device.Client(avtransport.Version1); ok && device.UDN == transportUDN {
					transport = avtransport.NewClient(soapClient)
				}
				if soapClient, ok := device.Client(contentdirectory.Version1); ok && device.UDN == directoryUDN {
					directory = contentdirectory.NewClient(soapClient)
				}
			}
			if transport == nil {
				http.Error(w, fmt.Sprintf("could not find AVTransport %s", transportUDN), http.StatusNotFound)
				return
			}
			if directory == nil {
				http.Error(w, fmt.Sprintf("could not find ContentDirectory %s", directoryUDN), http.StatusNotFound)
				return
			}

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
		})

	m.Path("/renderer/{udn}").
		Methods("POST").
		Queries("action", "play").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			udn := mustVar(r, "udn")

			ctx := r.Context()
			discoverCtx, _ := context.WithTimeout(ctx, 2*time.Second)
			devices, _, err := ssdp.Discover(discoverCtx, ssdp.URN(udn))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for _, device := range devices {
				if soapClient, ok := device.Client(avtransport.Version1); ok && device.UDN == udn {
					transport := avtransport.NewClient(soapClient)
					if err := transport.Play(ctx); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
			}
			http.Error(w, fmt.Sprintf("could not find renderer %s", udn), http.StatusNotFound)
		})

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

func maybeRedirectAfter(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)
		if redirect := r.Form.Get("redirect"); redirect != "" {
			http.Redirect(w, r, redirect, http.StatusFound)
		}
	}
}
