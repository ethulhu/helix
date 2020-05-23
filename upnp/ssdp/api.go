// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

// Package ssdp implements the Simple Service Discovery Protocol.
package ssdp

import "net/url"

type (
	// Device is an SSDP device.
	Device struct {
		// Name is the "friendly name" of a UPnP device.
		Name string

		// UDN is a unique identifier that can be used to rediscover a device.
		UDN string

		services map[URN]*url.URL
	}

	// URN is a UPnP service URN.
	URN string
)

const (
	RootDevice = URN("upnp:rootdevice")
	All        = URN("ssdp:all")
)
