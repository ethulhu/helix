package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ethulhu/helix/upnp/ssdp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
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
	args := struct {
		Items []upnpav.Item
		Queue queueArgs
	}{queue.Queue(), getQueueArgs()}

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
	log.Printf("setting transport: %v", udn)

	if err := queue.SetTransport(device); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func playQueue(w http.ResponseWriter, r *http.Request) {
	log.Print("playing")
	queue.Play()
}
func pauseQueue(w http.ResponseWriter, r *http.Request) {
	log.Print("pausing")
	queue.Pause()
}
func stopQueue(w http.ResponseWriter, r *http.Request) {
	log.Print("stopping")
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
	didl, err := directory.BrowseMetadata(ctx, upnpav.Object(object))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("could not add %v to queue: %v", object, err)
		return
	}

	if len(didl.Items) != 0 {
		for _, item := range didl.Items {
			queue.AddLast(item)
		}
	} else if len(didl.Containers) == 1 {
		didl, err := directory.BrowseChildren(ctx, upnpav.Object(object))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("could not add %v to queue: %v", object, err)
			return
		}
		for _, item := range didl.Items {
			queue.AddLast(item)
		}
	} else {
		http.Error(w, "could not find any items", http.StatusNotFound)
	}
}
func removeAllFromQueue(w http.ResponseWriter, r *http.Request) {
	queue.Clear()
}
