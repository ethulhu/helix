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

var (
	ErrNoSuchObject                    = upnpav.Error{Code: 701, Description: "No such object"}
	ErrInvalidCurrentTag               = upnpav.Error{Code: 702, Description: "Invalid CurrentTagValue"}
	ErrInvalidNewTag                   = upnpav.Error{Code: 703, Description: "Invalid NewTagValue"}
	ErrRequiredTag                     = upnpav.Error{Code: 704, Description: "Required tag"}
	ErrReadOnlyTag                     = upnpav.Error{Code: 705, Description: "Read Only tag"}
	ErrParameterMismatch               = upnpav.Error{Code: 706, Description: "Parameter mismatch"}
	ErrInvalidSearchCriteria           = upnpav.Error{Code: 708, Description: "Unsupported or invalid search criteria"}
	ErrInvalidSortCriteria             = upnpav.Error{Code: 709, Description: "Unsupported or invalid sort criteria"}
	ErrNoSuchContainer                 = upnpav.Error{Code: 710, Description: "No such container"}
	ErrRestrictedObject                = upnpav.Error{Code: 711, Description: "Restricted object"}
	ErrBadMetadata                     = upnpav.Error{Code: 712, Description: "Bad metadata"}
	ErrRestrictedParent                = upnpav.Error{Code: 713, Description: "Restricted parent object"}
	ErrNoSuchResource                  = upnpav.Error{Code: 714, Description: "No such resource"}
	ErrSourceResourceAccessDenied      = upnpav.Error{Code: 715, Description: "Resource access denied"}
	ErrTransferBusy                    = upnpav.Error{Code: 716, Description: "Transfer busy"}
	ErrNoSuchTransfer                  = upnpav.Error{Code: 717, Description: "No such file transfer"}
	ErrNoSuchDestinationResource       = upnpav.Error{Code: 718, Description: "No such destination resource"}
	ErrDestinationResourceAccessDenied = upnpav.Error{Code: 718, Description: "Destination resource access denied"}
	ErrCannotProcessRequest            = upnpav.Error{Code: 720, Description: "Cannot process the request"}
)

var SCPD = scpd.Must(scpd.Merge(
	scpd.Must(scpd.FromAction(browse, browseRequest{}, browseResponse{})),
	scpd.Must(scpd.FromAction(getSearchCapabilities, getSearchCapabilitiesRequest{}, getSearchCapabilitiesResponse{})),
	scpd.Must(scpd.FromAction(getSortCapabilities, getSortCapabilitiesRequest{}, getSortCapabilitiesResponse{})),
	scpd.Must(scpd.FromAction(searchA, searchRequest{}, searchResponse{})),
	scpd.Must(scpd.FromAction(getSystemUpdateID, getSystemUpdateIDRequest{}, getSystemUpdateIDResponse{})),
))
