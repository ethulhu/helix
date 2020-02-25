package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

// Index.

func getIndexHTML(w http.ResponseWriter, r *http.Request) {
	directories := devices.DevicesByURN(contentdirectory.Version1)
	transports := devices.DevicesByURN(avtransport.Version1)

	args := struct {
		Directories []*ssdp.Device
		Transports  []*ssdp.Device
	}{directories, transports}
	if err := indexTmpl.Execute(w, args); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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
	device, _ := devices.DeviceByUDN(udn)

	ctx := r.Context()
	didl, err := directory.Browse(ctx, contentdirectory.BrowseChildren, upnpav.Object(object))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	transports := devices.DevicesByURN(avtransport.Version1)

	args := struct {
		DIDL       *upnpav.DIDL
		Directory  *ssdp.Device
		Transports []*ssdp.Device
	}{didl, device, transports}
	if err := browseTmpl.Execute(w, args); err != nil {
		log.Printf("error rendering %v: %v", r.URL.Path, err)
		return
	}
}

// AVTransport handlers.

func getTransports(w http.ResponseWriter, r *http.Request) {
	transports := devices.DevicesByURN(avtransport.Version1)

	if err := transportsTmpl.Execute(w, transports); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func getTransportHTML(w http.ResponseWriter, r *http.Request) {
	udn := mustVar(r, "udn")

	transport, ok := devices.AVTransportByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find AVTransport %s", udn), http.StatusNotFound)
		return
	}
	device, _ := devices.DeviceByUDN(udn)

	ctx := r.Context()
	state, err := transport.TransportInfo(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var didl *upnpav.DIDL
	if state == avtransport.StatePlaying || state == avtransport.StatePaused {
		var err error
		_, didl, _, _, err = transport.PositionInfo(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	args := struct {
		DIDL          *upnpav.DIDL
		Transport     *ssdp.Device
		PlaybackState avtransport.State
	}{didl, device, state}
	if err := transportTmpl.Execute(w, args); err != nil {
		log.Printf("error rendering %v: %v", r.URL.Path, err)
		return
	}
}
func getTransportJSON(w http.ResponseWriter, r *http.Request) {
	udn := mustVar(r, "udn")

	transport, ok := devices.AVTransportByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find AVTransport %s", udn), http.StatusNotFound)
		return
	}
	device, _ := devices.DeviceByUDN(udn)

	ctx := r.Context()
	state, err := transport.TransportInfo(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var playing *playingStatus
	if state == avtransport.StatePlaying || state == avtransport.StatePaused {
		_, didl, elapsed, duration, err := transport.PositionInfo(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		playing = &playingStatus{
			Name:     didl.Items[0].Title,
			Duration: duration.Milliseconds(),
			Elapsed:  elapsed.Milliseconds(),
		}
	}

	data := transportStatus{
		UDN:     device.UDN,
		Name:    device.Name,
		State:   state,
		Playing: playing,
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("could not marshal transport JSON: %v", err)
	}
	w.Write(bytes)
}

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

	_ = transport.Stop(ctx)
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
