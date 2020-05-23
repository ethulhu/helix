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

type (
	manifestXML struct {
		XMLName     xml.Name `xml:"urn:schemas-upnp-org:device-1-0 root"`
		SpecVersion struct {
			Major int `xml:"specVersion>major"`
			Minor int `xml:"specVersion>minor"`
		}
		Device deviceXML `xml:"device"`
	}
	deviceXML struct {
		DeviceType   string `xml:"deviceType"`
		FriendlyName string `xml:"friendlyName"`
		UDN          string `xml:"UDN"`

		Devices         []deviceXML  `xml:"deviceList>device"`
		Icons           []iconXML    `xml:"iconList>icon"`
		Services        []serviceXML `xml:"serviceList>service"`
		PresentationURL string       `xml:"presentationURL"`
	}
	iconXML struct {
		MIMEType string `xml:"mimetype"`
		Width    int    `xml:"width"`
		Height   int    `xml:"height"`
		Depth    int    `xml:"depth"`
		URL      string `xml:"url"`
	}
	serviceXML struct {
		ServiceType URN    `xml:"serviceType"`
		ServiceID   string `xml:"serviceId"`
		ControlURL  string `xml:"controlURL"`
		EventSubURL string `xml:"eventSubURL"`
		SCPDURL     string `xml:"SCPDURL"`
	}
)

func newDevice(manifestURL *url.URL, rawManifest []byte) (*Device, error) {
	m, err := parseDeviceManifest(rawManifest)
	if err != nil {
		return nil, fmt.Errorf("could not parse device manifest: %w", err)
	}

	services := map[URN]*url.URL{}
	for _, s := range m.Device.Services {
		// TODO: how do ServiceType and ServiceID differ?
		serviceURL := *manifestURL
		serviceURL.Path = s.ControlURL
		services[s.ServiceType] = &serviceURL
	}

	return &Device{
		Name:     m.Device.FriendlyName,
		UDN:      m.Device.UDN,
		services: services,
	}, nil
}

// SOAPClient returns a SOAP client for the given URN, and whether or not that client exists.
// A nil Device always returns (nil, false).
func (d *Device) SOAPClient(urn URN) (soap.Client, bool) {
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

func parseDeviceManifest(raw []byte) (manifestXML, error) {
	m := manifestXML{}
	err := xml.Unmarshal(raw, &m)
	return m, err
}
