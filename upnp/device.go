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

	log "github.com/sirupsen/logrus"
)

type (
	service struct {
		SCPD      scpd.Document
		Interface soap.Interface
		ID        ServiceID
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
	d.Manufacturer = d.Manufacturer
	d.ManufacturerURL = d.ManufacturerURL
	d.ModelDescription = d.ModelDescription
	d.ModelName = d.ModelName
	d.ModelNumber = d.ModelNumber
	d.ModelURL = d.ModelURL
	d.SerialNumber = d.SerialNumber

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
func (d *Device) allURNs() []URN {
	return append(d.Services(), URN(d.DeviceType), RootDevice)
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
	fields := log.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
		"remote": r.RemoteAddr,
	}

	if r.URL.Path == "/" {
		bytes, err := xml.Marshal(d.manifest())
		if err != nil {
			panic(fmt.Sprintf("could not marshal manifest: %v", err))
		}
		fmt.Fprint(w, xml.Header)
		w.Write(bytes)
		log.WithFields(fields).Info("served SSDP manifest")
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
			fmt.Fprint(w, xml.Header)
			w.Write(bytes)
			log.WithFields(fields).Info("served SCPD")
			return

		case "POST":
			log.WithFields(fields).Info("serving SOAP")
			soap.Handle(w, r, service.Interface)
			return
		}
	}

	fields["error"] = "not found"
	log.WithFields(fields).Error("not found")
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
		},
	}

	for urn, service := range d.serviceByURN {
		doc.Device.Services = append(doc.Device.Services, ssdp.Service{
			ServiceType: string(urn),
			ServiceID:   string(service.ID),
			SCPDURL:     "/" + string(urn),
			ControlURL:  "/" + string(urn),
			EventSubURL: "/" + string(urn),
		})
	}

	for _, icon := range d.Icons {
		doc.Device.Icons = append(doc.Device.Icons, icon.ssdpIcon())
	}

	return doc
}

func (d *Device) Handle(urn URN, id ServiceID, doc scpd.Document, handler soap.Interface) {
	d.serviceByURN[urn] = service{
		ID:        id,
		Interface: handler,
		SCPD:      doc,
	}
}