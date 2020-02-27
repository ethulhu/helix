package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
)

func getQueueJSON(w http.ResponseWriter, r *http.Request) {
	data := struct {
		UDN   string            `json:"udn,omitempty"`
		Name  string            `json:"name,omitempty"`
		State avtransport.State `json:"state"`
		Items []upnpav.Item     `json:"items,omitempty"`
	}{queue.UDN(), queue.Name(), queue.State(), queue.Queue()}

	bytes, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("could not marshal /queue JSON: %v", err))
	}
	w.Write(bytes)
}
func getQueueHTML(w http.ResponseWriter, r *http.Request) {
	transports := devices.DevicesByURN(avtransport.Version1)

	udn := "none"
	if queue.UDN() != "" {
		udn = queue.UDN()
	}

	args := struct {
		CurrentUDN  string
		State       avtransport.State
		Items       []upnpav.Item
		Transports  []*ssdp.Device
	}{udn, queue.State(), queue.Queue(), transports}

	if err := queueTmpl.Execute(w, args); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func setQueueTransport(w http.ResponseWriter, r *http.Request) {
	udn := mustVar(r, "transport")

	// "none" is a magic value to unset the transport.
	var device *ssdp.Device
	if udn != "none" {
		var ok bool
		device, ok = devices.DeviceByUDN(udn)
		if !ok {
			http.Error(w, fmt.Sprintf("could not find device %s", udn), http.StatusNotFound)
			return
		}
	}

	if err := queue.SetTransport(device); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func playQueue(w http.ResponseWriter, r *http.Request) {
	if queue.UDN() == "" {
		transports := devices.DevicesByURN(avtransport.Version1)
		if len(transports) == 1 {
			queue.SetTransport(transports[0])
		}
	}
	queue.Play()
}
func pauseQueue(w http.ResponseWriter, r *http.Request) {
	queue.Pause()
}
func stopQueue(w http.ResponseWriter, r *http.Request) {
	queue.Stop()
}
func addObjectToQueue(w http.ResponseWriter, r *http.Request) {
	udn := mustVar(r, "directory")
	object := mustVar(r, "object")

	directory, ok := devices.ContentDirectoryByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("could not find ContentDirectory %s", udn), http.StatusNotFound)
		return
	}

	ctx := r.Context()
	didl, err := directory.Browse(ctx, contentdirectory.BrowseMetadata, upnpav.Object(object))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	queue.AddLast(didl.Items[0])
}
func removeAllFromQueue(w http.ResponseWriter, r *http.Request) {
	queue.Clear()
}
