// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package contentdirectory

import (
	"encoding/xml"

	"github.com/ethulhu/helix/upnpav"
	"github.com/ethulhu/helix/upnpav/contentdirectory/search"
)

type (
	getSearchCapabilitiesResponse struct {
		XMLName      xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 GetSearchCapabilitiesResponse"`
		Capabilities string   `xml:"SearchCaps"`
	}

	browseFlag    string
	browseRequest struct {
		XMLName xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 Browse"`

		// Object is the ID of the Object being browsed.
		// An ObjectID value of 0 corresponds to the root object of the Content Directory.
		Object upnpav.ObjectID `xml:"ObjectID"`

		// BrowseFlag specifies whether to return data about Object or Object's children.
		BrowseFlag browseFlag `xml:"BrowseFlag"`

		// Filter is a comma-separated list of properties (e.g. "upnp:artist"), or "*".
		Filter string `xml:"Filter"`

		// StartingIndex is a zero-based offset to enumerate children under the container specified by ObjectID.
		// Must be 0 if BrowseFlag is equal to BrowseMetadata.
		StartingIndex string `xml:"StartingIndex"`

		// Requested number of entries under the object specified by ObjectID.
		// RequestedCount =0 indicates request all entries.
		RequestedCount int `xml:"RequestedCount"`

		// SortCriteria is a comma-separated list of "signed" properties.
		// For example "+upnp:artist" means "return objects sorted ascending by artist",
		// and "+upnp:artist,-dc:date" means "return objects sorted by (ascending artist, descending date)".
		SortCriteria string `xml:"SortCriteria"`
	}
	browseResponse struct {
		// Result is a DIDL-Lite XML document.
		Result []byte `xml:"Result"`

		// Number of objects returned in this result.
		// If BrowseMetadata is specified in the BrowseFlags, then NumberReturned = 1
		NumberReturned int `xml:"NumberReturned"`

		TotalMatches int `xml:"TotalMatches"`

		UpdateID string `xml:"UpdateID"`

		// Contents is the contents of the response, for debugging use.
		Contents []byte `xml:",innerxml"`
	}

	searchRequest struct {
		Object upnpav.ObjectID `xml:"ObjectID"`

		// SearchCriteria is a dirty hack,
		// because encoding/xml will escape the string,
		// and we need it to not.
		SearchCriteria struct {
			Criteria search.Criteria `xml:",innerxml"`
		} `xml:"SearchCriteriardata"`

		Filter         string `xml:"Filter"`
		StartingIndex  string `xml:"StartingIndex"`
		RequestedCount int    `xml:"RequestedCount"`
		SortCriteria   string `xml:"SortCriteria"`
	}
	searchResponse struct {
		// Result is a DIDL-Lite XML document.
		Result []byte `xml:"Result"`

		NumberReturned string `xml:"NumberReturned"`
		TotalMatches   string `xml:"TotalMatches"`
		UpdateID       string `xml:"UpdateID"`
	}
)

const (
	browseMetadata = browseFlag("BrowseMetadata")
	browseChildren = browseFlag("BrowseDirectChildren")
)
