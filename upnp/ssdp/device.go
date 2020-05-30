// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package ssdp

import (
	"encoding/xml"
	"fmt"
	"net/url"

	"github.com/ethulhu/helix/soap"
)

func newDeviceDeprecated(manifestURL *url.URL, rawManifest []byte) (*DeviceDeprecated, error) {
	m := Document{}
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

	return &DeviceDeprecated{
		Name:     m.Device.FriendlyName,
		UDN:      m.Device.UDN,
		services: services,
	}, nil
}

// SOAPInterface returns a SOAP interface for the given URN, and whether or not that interface exists.
// A nil DeviceDeprecated always returns (nil, false).
func (d *DeviceDeprecated) SOAPInterface(urn URN) (soap.Interface, bool) {
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
// A nil DeviceDeprecated always returns nil.
func (d *DeviceDeprecated) Services() []URN {
	if d == nil {
		return nil
	}

	var urns []URN
	for urn := range d.services {
		urns = append(urns, urn)
	}
	return urns
}
