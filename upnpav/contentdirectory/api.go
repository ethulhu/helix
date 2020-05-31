// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package contentdirectory

import (
	"context"

	"github.com/ethulhu/helix/upnp"
	"github.com/ethulhu/helix/upnp/scpd"
	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
)

type (
	Interface interface {
		// SearchCapabilities returns the search capabilities of the ContentDirectory service.
		SearchCapabilities(context.Context) ([]string, error)

		// SortCapabilities returns the sort capabilities of the ContentDirectory service.
		SortCapabilities(context.Context) ([]string, error)

		// BrowseMetadata shows information about a given object.
		BrowseMetadata(context.Context, upnpav.ObjectID) (*upnpav.DIDLLite, error)

		// BrowseChildren lists the child objects of a given object.
		BrowseChildren(context.Context, upnpav.ObjectID) (*upnpav.DIDLLite, error)

		// Search queries the ContentDirectory service for objects under a given object that match a given criteria.
		Search(context.Context, upnpav.ObjectID, search.Criteria) (*upnpav.DIDLLite, error)

		SystemUpdateID(ctx context.Context) (uint, error)
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

var SCPD = scpd.Must(scpd.Merge(
	scpd.Must(scpd.FromAction(browse, browseRequest{}, browseResponse{})),
	scpd.Must(scpd.FromAction(getSearchCapabilities, getSearchCapabilitiesRequest{}, getSearchCapabilitiesResponse{})),
	scpd.Must(scpd.FromAction(getSortCapabilities, getSortCapabilitiesRequest{}, getSortCapabilitiesResponse{})),
	scpd.Must(scpd.FromAction(searchA, searchRequest{}, searchResponse{})),
	scpd.Must(scpd.FromAction(getSystemUpdateID, getSystemUpdateIDRequest{}, getSystemUpdateIDResponse{})),
))