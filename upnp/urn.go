// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package upnp

type (
	// URN is a UPnP service URN.
	URN string

	ServiceID  string
	DeviceType string
)

const (
	RootDevice = URN("upnp:rootdevice")
	All        = URN("ssdp:all")
)
