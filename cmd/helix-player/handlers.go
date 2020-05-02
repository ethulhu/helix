package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/ethulhu/helix/httputil"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/avtransport"
	"github.com/ethulhu/helix/upnpav/contentdirectory"
	"github.com/gorilla/mux"
)

// ContentDirectory handlers.

func getDirectoriesJSON(w http.ResponseWriter, r *http.Request) {
	devices := directories.Devices()

	data := []directory{}
	for _, device := range devices {
		data = append(data, directoryFromDevice(device))
	}

	httputil.MustWriteJSON(w, data)
}

func getDirectoryJSON(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]

	device, ok := directories.DeviceByUDN(udn)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown ContentDirectory: %v", udn), http.StatusNotFound)
		return
	}

	data := directoryFromDevice(device)

	httputil.MustWriteJSON(w, data)
}

func getObjectJSON(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]
	objectID := mux.Vars(r)["object"]

	device, _ := directories.DeviceByUDN(udn)
	client, ok := device.SOAPClient(contentdirectory.Version1)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown ContentDirectory: %s", udn), http.StatusNotFound)
		return
	}
	directory := contentdirectory.NewClient(client)

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

	httputil.MustWriteJSON(w, data)
}

func getObjectByType(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]
	object := mux.Vars(r)["object"]
	mimetypeRaw := mux.Vars(r)["mimetype"]

	log.Printf("%v udn %q object %q MIME-type %q", r.Method, udn, object, mimetypeRaw)

	mimetype, _, err := mime.ParseMediaType(mimetypeRaw)
	mimeParts := strings.Split(mimetype, "/")
	if err != nil || len(mimeParts) != 2 {
		http.Error(w, fmt.Sprintf("invalid MIME-Type %q: %v", mimetypeRaw, err), http.StatusBadRequest)
		return
	}

	device, _ := directories.DeviceByUDN(udn)
	client, ok := device.SOAPClient(contentdirectory.Version1)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown ContentDirectory: %s", udn), http.StatusNotFound)
		return
	}
	directory := contentdirectory.NewClient(client)

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

// AVTransport handlers.

func getTransportsJSON(w http.ResponseWriter, r *http.Request) {
	devices := transports.Devices()

	data := []transport{}
	ctx := r.Context()
	for _, device := range devices {
		client, ok := device.SOAPClient(avtransport.Version1)
		if !ok {
			continue
		}
		transport := avtransport.NewClient(client)
		state, _, err := transport.TransportInfo(ctx)
		if err != nil {
			continue
		}
		data = append(data, transportFromDeviceAndInfo(device, state))
	}

	httputil.MustWriteJSON(w, data)
}
func getTransportJSON(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]

	device, _ := transports.DeviceByUDN(udn)
	client, ok := device.SOAPClient(avtransport.Version1)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown AVTransport: %v", udn), http.StatusNotFound)
		return
	}
	transport := avtransport.NewClient(client)

	ctx := r.Context()
	state, _, err := transport.TransportInfo(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get status from AVTransport: %v", err), http.StatusInternalServerError)
		return
	}

	data := transportFromDeviceAndInfo(device, state)

	httputil.MustWriteJSON(w, data)
}

func playTransport(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]

	device, _ := transports.DeviceByUDN(udn)
	client, ok := device.SOAPClient(avtransport.Version1)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown AVTransport: %v", udn), http.StatusNotFound)
		return
	}
	transport := avtransport.NewClient(client)

	ctx := r.Context()
	if err := transport.Play(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func pauseTransport(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]

	device, _ := transports.DeviceByUDN(udn)
	client, ok := device.SOAPClient(avtransport.Version1)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown AVTransport: %v", udn), http.StatusNotFound)
		return
	}
	transport := avtransport.NewClient(client)

	ctx := r.Context()
	if err := transport.Pause(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func stopTransport(w http.ResponseWriter, r *http.Request) {
	udn := mux.Vars(r)["udn"]

	device, _ := transports.DeviceByUDN(udn)
	client, ok := device.SOAPClient(avtransport.Version1)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown AVTransport: %v", udn), http.StatusNotFound)
		return
	}
	transport := avtransport.NewClient(client)

	ctx := r.Context()
	if err := transport.Stop(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
