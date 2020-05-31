// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package contentdirectory

import (
	"context"

	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
)

type (
	Interface interface {
		// BrowseMetadata shows information about a given object.
		BrowseMetadata(context.Context, upnpav.ObjectID) (*upnpav.DIDLLite, error)

		// BrowseChildren lists the child objects of a given object.
		BrowseChildren(context.Context, upnpav.ObjectID) (*upnpav.DIDLLite, error)

		// SearchCapabilities returns the search capabilities of the ContentDirectory service.
		SearchCapabilities(context.Context) ([]string, error)

		// Search queries the ContentDirectory service for objects under a given object that match a given criteria.
		Search(context.Context, upnpav.ObjectID, search.Criteria) (*upnpav.DIDLLite, error)
	}
)

const (
	Version1   = upnp.URN("urn:schemas-upnp-org:service:ContentDirectory:1")
	Version2   = upnp.URN("urn:schemas-upnp-org:service:ContentDirectory:2")
	Version3   = upnp.URN("urn:schemas-upnp-org:service:ContentDirectory:3")
	ServiceID  = upnp.ServiceID("urn:upnp-org:serviceId:ContentDirectory")
	DeviceType = upnp.DeviceType("urn:schemas-upnp-org:device:MediaServer:1")
)

const (
	Root = upnpav.ObjectID("0")
)
