package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

// ContentDirectory handlers.

func getDirectories(w http.ResponseWriter, r *http.Request) {

	var ds []*ssdp.Device
	devicesLock.Lock()
	for _, device := range devices {
		if _, ok := device.Client(contentdirectory.Version1); ok {
			ds = append(ds, device)
		}
	}
	devicesLock.Unlock()

	if err := directoriesTmpl.Execute(w, devices); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func getDirectory(w http.ResponseWriter, r *http.Request) {
	udn := mustVar(r, "udn")
	http.Redirect(w, r, fmt.Sprintf("/browse/%v/0", udn), http.StatusFound)
}
func getObject(w http.ResponseWriter, r *http.Request) {
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
}

// AVTransport handlers.

func play(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	transport := ctx.Value("AVTransport").(avtransport.Client)
	if err := transport.Play(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maybeRedirect(w, r)
}
func pause(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	transport := ctx.Value("AVTransport").(avtransport.Client)
	if err := transport.Pause(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maybeRedirect(w, r)
}
func stop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	transport := ctx.Value("AVTransport").(avtransport.Client)
	if err := transport.Stop(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maybeRedirect(w, r)
}
func playObject(w http.ResponseWriter, r *http.Request) {
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
}
