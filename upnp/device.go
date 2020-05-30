// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnp

import (
	"encoding/xml"
	"fmt"
	"net/url"

	"github.com/ethulhu/helix/soap"
	"github.com/ethulhu/helix/upnp/ssdp"
)

type (
	// Device is an SSDP device that you shouldn't use.
	Device struct {
		// Name is the "friendly name" of a UPnP device.
		Name string

		// UDN is a unique identifier that can be used to rediscover a device.
		UDN string

		services map[URN]*url.URL
	}
)

func newDevice(manifestURL *url.URL, rawManifest []byte) (*Device, error) {
	m := ssdp.Document{}
	if err := xml.Unmarshal(rawManifest, &m); err != nil {
		return nil, fmt.Errorf("could not parse device manifest: %w", err)
	}

	services := map[URN]*url.URL{}
	for _, s := range m.Device.Services {
		// TODO: how do ServiceType and ServiceID differ?
		serviceURL := *manifestURL
		serviceURL.Path = s.ControlURL
		services[URN(s.ServiceType)] = &serviceURL
	}

	return &Device{
		Name:     m.Device.FriendlyName,
		UDN:      m.Device.UDN,
		services: services,
	}, nil
}

// SOAPInterface returns a SOAP interface for the given URN, and whether or not that interface exists.
// A nil Device always returns (nil, false).
func (d *Device) SOAPInterface(urn URN) (soap.Interface, bool) {
	if d == nil {
		return nil, false
	}

	baseURL, ok := d.services[urn]
	if !ok {
		return nil, false
	}
	return soap.NewClient(baseURL), true
}

// Services lists URNs advertised by the device.
// A nil Device always returns nil.
func (d *Device) Services() []URN {
	if d == nil {
		return nil
	}

	var urns []URN
	for urn := range d.services {
		urns = append(urns, urn)
	}
	return urns
}