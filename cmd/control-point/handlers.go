package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

// ContentDirectory handlers.

func getDirectories(w http.ResponseWriter, r *http.Request) {
	directories := devices.DevicesByURN(contentdirectory.Version1)

	if err := directoriesTmpl.Execute(w, directories); err != nil {
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

	directory, ok := devices.ContentDirectoryByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find ContentDirectory %s", udn), http.StatusNotFound)
		return
	}

	ctx := r.Context()
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
	udn := mustVar(r, "udn")

	transport, ok := devices.AVTransportByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find AVTransport %s", udn), http.StatusNotFound)
		return
	}

	ctx := r.Context()
	if err := transport.Play(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maybeRedirect(w, r)
}
func pause(w http.ResponseWriter, r *http.Request) {
	udn := mustVar(r, "udn")

	transport, ok := devices.AVTransportByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find AVTransport %s", udn), http.StatusNotFound)
		return
	}

	ctx := r.Context()
	if err := transport.Pause(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maybeRedirect(w, r)
}
func stop(w http.ResponseWriter, r *http.Request) {
	udn := mustVar(r, "udn")

	transport, ok := devices.AVTransportByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find AVTransport %s", udn), http.StatusNotFound)
		return
	}

	ctx := r.Context()
	if err := transport.Stop(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maybeRedirect(w, r)
}
func playObject(w http.ResponseWriter, r *http.Request) {
	directoryUDN := mustVar(r, "directory")
	transportUDN := mustVar(r, "udn")
	object := mustVar(r, "object")

	directory, ok := devices.ContentDirectoryByUDN(directoryUDN)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find ContentDirectory %s", directoryUDN), http.StatusNotFound)
		return
	}
	transport, ok := devices.AVTransportByUDN(transportUDN)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find AVTransport %s", transportUDN), http.StatusNotFound)
		return
	}

	ctx := r.Context()
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
