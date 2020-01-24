// Package ssdp implements the Simple Service Discovery Protocol.
package ssdp

import "github.com/ethulhu/helix/soap"

type (
	// Device is an SSDP device.
	Device interface {
		// Name is the "friendly name" of a UPnP device.
		Name() string
		// Client returns a SOAP client for the given URN, and whether or not that client exists.
		Client(URN) (soap.Client, bool)
		// Services lists URNs advertised by the device.
		Services() []URN
	}

	// URN is a UPnP service URN.
	URN string
)

const (
	RootDevice = URN("upnp:rootdevice")
	All        = URN("ssdp:all")
)
