package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

// Index.

type queueArgs struct {
	Transports []*ssdp.Device
	CurrentUDN string
	Items      []upnpav.Item
}

func getQueueArgs() queueArgs {
	transports := devices.DevicesByURN(avtransport.Version1)
	sort.Slice(transports, func(i, j int) bool {
		return transports[i].Name < transports[j].Name
	})

	udn := "none"
	if queue.UDN() != "" {
		udn = queue.UDN()
	}

	return queueArgs{
		Transports: transports,
		CurrentUDN: udn,
		Items:      queue.Queue(),
	}
}

func getIndexHTML(w http.ResponseWriter, r *http.Request) {
	directories := devices.DevicesByURN(contentdirectory.Version1)
	sort.Slice(directories, func(i, j int) bool {
		return directories[i].Name < directories[j].Name
	})

	args := struct {
		Directories []*ssdp.Device
		Queue       queueArgs
	}{directories, getQueueArgs()}
	if err := indexTmpl.Execute(w, args); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("could not render %v: %v", r.URL.Path, err)
	}
}

// ContentDirectory handlers.

func getDirectoriesHTML(w http.ResponseWriter, r *http.Request) {
	directories := devices.DevicesByURN(contentdirectory.Version1)
	sort.Slice(directories, func(i, j int) bool {
		return directories[i].Name < directories[j].Name
	})

	args := struct {
		Directories []*ssdp.Device
		Queue       queueArgs
	}{directories, getQueueArgs()}
	if err := directoriesTmpl.Execute(w, args); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func getDirectory(w http.ResponseWriter, r *http.Request) {
	udn := mustVar(r, "udn")
	http.Redirect(w, r, fmt.Sprintf("/browse/%v/0", udn), http.StatusFound)
}
func getObjectHTML(w http.ResponseWriter, r *http.Request) {
	object := mustVar(r, "object")
	udn := mustVar(r, "udn")

	directory, ok := devices.ContentDirectoryByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find ContentDirectory %s", udn), http.StatusNotFound)
		return
	}
	device, _ := devices.DeviceByUDN(udn)

	ctx := r.Context()
	self, err := directory.BrowseMetadata(ctx, upnpav.Object(object))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: support browsing non-container objects.
	if len(self.Containers) == 0 {
		http.Error(w, "could not find container", http.StatusNotFound)
	}

	children, err := directory.BrowseChildren(ctx, upnpav.Object(object))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	args := struct {
		Container upnpav.Container
		DIDL      *upnpav.DIDL
		Directory *ssdp.Device
		Queue     queueArgs
	}{self.Containers[0], children, device, getQueueArgs()}
	if err := browseTmpl.Execute(w, args); err != nil {
		log.Printf("error rendering %v: %v", r.URL.Path, err)
		return
	}
}

// AVTransport handlers.

func getTransports(w http.ResponseWriter, r *http.Request) {
	args := struct {
		Queue queueArgs
	}{getQueueArgs()}
	if err := transportsTmpl.Execute(w, args); err != nil {
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
	state, _, err := transport.TransportInfo(ctx)
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
		Queue         queueArgs
	}{didl, device, state, getQueueArgs()}
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
	state, _, err := transport.TransportInfo(ctx)
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
	didl, err := directory.BrowseMetadata(ctx, upnpav.Object(object))
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
}
