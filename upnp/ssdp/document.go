// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package ssdp

import "encoding/xml"

const xmlns = "urn:schemas-upnp-org:device-1-0"

type (
	Document struct {
		XMLName     xml.Name    `xml:"urn:schemas-upnp-org:device-1-0 root"`
		SpecVersion SpecVersion `xml:"specVersion"`
		Device      Device      `xml:"device"`
	}

	SpecVersion struct {
		Major int `xml:"major"`
		Minor int `xml:"minor"`
	}

	Device struct {
		DeviceType       string `xml:"deviceType,omitempty"`
		FriendlyName     string `xml:"friendlyName,omitempty"`
		Manufacturer     string `xml:"manufacturer,omitempty"`
		ManufacturerURL  string `xml:"manufacturerURL,omitempty"`
		ModelDescription string `xml:"modelDescription,omitempty"`
		ModelName        string `xml:"modelName,omitempty"`
		ModelNumber      string `xml:"modelNumber,omitempty"`
		ModelURL         string `xml:"modelURL,omitempty"`
		SerialNumber     string `xml:"serialNumber,omitempty"`
		UDN              string `xml:"UDN,omitempty"`

		Devices         []Device  `xml:"deviceList>device"`
		Icons           []Icon    `xml:"iconList>icon"`
		Services        []Service `xml:"serviceList>service"`
		PresentationURL string    `xml:"presentationURL,omitempty"`
	}
	Icon struct {
		MIMEType string `xml:"mimetype"`
		Width    int    `xml:"width"`
		Height   int    `xml:"height"`
		Depth    int    `xml:"depth"`
		URL      string `xml:"url"`
	}
	Service struct {
		ServiceType string `xml:"serviceType"` // URN
		ServiceID   string `xml:"serviceId"`
		ControlURL  string `xml:"controlURL"`
		EventSubURL string `xml:"eventSubURL"`
		SCPDURL     string `xml:"SCPDURL"`
	}
)