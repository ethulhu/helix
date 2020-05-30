// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnp

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnp/scpd"
	"github.com/ethulhu/helix/upnp/ssdp"
)

type (
	service struct {
		SCPD      scpd.Document
		Interface soap.Interface
	}

	// Device is an UPnP device.
	Device struct {
		// Name is the "friendly name" of a UPnP device.
		Name string

		// UDN is a unique identifier that can be used to rediscover a device.
		UDN string

		serviceByURN map[URN]service
	}
)

func NewDevice(name, udn string) *Device {
	return &Device{
		Name:         name,
		UDN:          udn,
		serviceByURN: map[URN]service{},
	}
}

func newDevice(manifestURL *url.URL, manifest ssdp.Document) (*Device, error) {
	d := NewDevice(manifest.Device.FriendlyName, manifest.Device.UDN)
	for _, s := range manifest.Device.Services {
		// TODO: get the actual SCPD.
		serviceURL := *manifestURL
		serviceURL.Path = s.ControlURL
		d.serviceByURN[URN(s.ServiceType)] = service{
			Interface: soap.NewClient(&serviceURL),
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

	var urns []URN
	for urn := range d.serviceByURN {
		urns = append(urns, urn)
	}
	return urns
}

// SOAPClient returns a SOAP client for the given URN, and whether or not that client exists.
// A nil Device always returns (nil, false).
func (d *Device) SOAPInterface(urn URN) (soap.Interface, bool) {
	if d == nil {
		return nil, false
	}

	service, ok := d.serviceByURN[urn]
	if !ok {
		return nil, false
	}
	return service.Interface, true
}

// ServeHTTP serves the SSDP/SCPD UPnP discovery interface, and marshals SOAP requests.
func (d *Device) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		bytes, err := xml.Marshal(d.manifest())
		if err != nil {
			panic(fmt.Sprintf("could not marshal manifest: %v", err))
		}
		w.Write(bytes)
		return
	}

	urn := URN(r.URL.Path[1:])
	if service, ok := d.serviceByURN[urn]; ok {
		switch r.Method {
		case "GET":
			bytes, err := xml.Marshal(service.SCPD)
			if err != nil {
				panic(fmt.Sprintf("could not marshal SCPD for %v: %v", urn, err))
			}
			w.Write(bytes)
			return
		case "POST":
			// TODO: handle SOAP calls
			// soap.Call(w, r, service.Interface)
		}
	}

	http.NotFound(w, r)
}

func (d *Device) manifest() ssdp.Document {
	doc := ssdp.Document{
		SpecVersion: ssdp.SpecVersion{
			Major: 1,
			Minor: 1,
		},
		Device: ssdp.Device{
			FriendlyName: d.Name,
			UDN:          d.UDN,
		},
	}

	for urn := range d.serviceByURN {
		doc.Device.Services = append(doc.Device.Services, ssdp.Service{
			ServiceType: string(urn),
			SCPDURL:     "/" + string(urn),
		})
	}

	return doc
}

func (d *Device) HandleURN(urn URN, doc scpd.Document, handler soap.Interface) {
	d.serviceByURN[urn] = service{
		SCPD:      doc,
		Interface: handler,
	}
}
