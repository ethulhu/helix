// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnp

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/ethulhu/helix/logger"
	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnp/scpd"
	"github.com/ethulhu/helix/upnp/ssdp"
)

type (
	service struct {
		SCPD          scpd.Document
		SOAPInterface soap.Interface
		ID            ServiceID
	}

	// Device is an UPnP device.
	Device struct {
		// Name is the "friendly name" of a UPnP device.
		Name string

		// UDN is a unique identifier that can be used to rediscover a device.
		UDN string

		// DeviceType is yet another URN-alike.
		DeviceType DeviceType

		Icons []Icon

		// Below are optional metadata fields.

		Manufacturer     string
		ManufacturerURL  string
		ModelDescription string
		ModelName        string
		ModelNumber      string
		ModelURL         string
		SerialNumber     string

		PresentationURL string

		mu           sync.RWMutex
		serviceByURN map[URN]service
	}
)

func newDevice(manifestURL *url.URL, manifest ssdp.Document) (*Device, error) {
	d := &Device{
		Name:             manifest.Device.FriendlyName,
		UDN:              manifest.Device.UDN,
		Manufacturer:     manifest.Device.Manufacturer,
		ManufacturerURL:  manifest.Device.ManufacturerURL,
		ModelDescription: manifest.Device.ModelDescription,
		ModelName:        manifest.Device.ModelName,
		ModelNumber:      manifest.Device.ModelNumber,
		ModelURL:         manifest.Device.ModelURL,
		SerialNumber:     manifest.Device.SerialNumber,
	}

	if manifest.Device.PresentationURL != "" {
		presentationURL := *manifestURL
		presentationURL.Path = manifest.Device.PresentationURL
		d.PresentationURL = presentationURL.String()
	}

	d.serviceByURN = map[URN]service{}
	for _, s := range manifest.Device.Services {
		// TODO: get the actual SCPD.
		serviceURL := *manifestURL
		serviceURL.Path = s.ControlURL
		d.serviceByURN[URN(s.ServiceType)] = service{
			SOAPInterface: soap.NewClient(&serviceURL),
		}
	}
	return d, nil
}

// Services lists URNs advertised by the device.
// A nil Device always returns nil.
func (d *Device) Services() []URN {
	if d == nil {
		return nil
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	var urns []URN
	for urn := range d.serviceByURN {
		urns = append(urns, urn)
	}
	return urns
}
func (d *Device) allURNs() []URN {
	if d == nil {
		return nil
	}
	return append(d.Services(), URN(d.DeviceType), RootDevice)
}

// SOAPClient returns a SOAP client for the given URN, and whether or not that client exists.
// A nil Device always returns (nil, false).
func (d *Device) SOAPInterface(urn URN) (soap.Interface, bool) {
	if d == nil {
		return nil, false
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	service, ok := d.serviceByURN[urn]
	if !ok {
		return nil, false
	}
	return service.SOAPInterface, true
}

// ServeHTTP serves the SSDP/SCPD UPnP discovery interface, and marshals SOAP requests.
func (d *Device) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log, ctx := logger.FromContext(r.Context())
	r = r.WithContext(ctx)

	log.AddField("http.client", r.RemoteAddr)
	log.AddField("http.method", r.Method)
	log.AddField("http.path", r.URL.Path)

	if r.URL.Path == "/" {
		bytes, err := xml.Marshal(d.manifest())
		if err != nil {
			panic(fmt.Sprintf("could not marshal manifest: %v", err))
		}
		fmt.Fprint(w, xml.Header)
		w.Write(bytes)
		log.Info("served SSDP manifest")
		return
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	urn := URN(r.URL.Path[1:])
	if service, ok := d.serviceByURN[urn]; ok {
		switch r.Method {
		case "GET":
			bytes, err := xml.Marshal(service.SCPD)
			if err != nil {
				panic(fmt.Sprintf("could not marshal SCPD for %v: %v", urn, err))
			}
			fmt.Fprint(w, xml.Header)
			w.Write(bytes)
			log.Info("served SCPD")
			return

		case "POST":
			soap.Handle(w, r, service.SOAPInterface)
			return
		}
	}

	log.Warning("not found")
	http.NotFound(w, r)
}

func (d *Device) manifest() ssdp.Document {
	doc := ssdp.Document{
		SpecVersion: ssdp.Version,
		Device: ssdp.Device{
			DeviceType:   string(d.DeviceType),
			FriendlyName: d.Name,
			UDN:          d.UDN,

			Manufacturer:     d.Manufacturer,
			ManufacturerURL:  d.ManufacturerURL,
			ModelDescription: d.ModelDescription,
			ModelName:        d.ModelName,
			ModelNumber:      d.ModelNumber,
			ModelURL:         d.ModelURL,
			SerialNumber:     d.SerialNumber,

			PresentationURL: d.PresentationURL,
		},
	}

	for _, icon := range d.Icons {
		doc.Device.Icons = append(doc.Device.Icons, icon.ssdpIcon())
	}

	d.mu.RLock()
	defer d.mu.RUnlock()
	for urn, service := range d.serviceByURN {
		doc.Device.Services = append(doc.Device.Services, ssdp.Service{
			ServiceType: string(urn),
			ServiceID:   string(service.ID),
			SCPDURL:     "/" + string(urn),
			ControlURL:  "/" + string(urn),
			EventSubURL: "/" + string(urn),
		})
	}

	return doc
}

func (d *Device) Handle(urn URN, id ServiceID, doc scpd.Document, handler soap.Interface) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.serviceByURN == nil {
		d.serviceByURN = map[URN]service{}
	}

	d.serviceByURN[urn] = service{
		ID:            id,
		SOAPInterface: handler,
		SCPD:          doc,
	}
}
